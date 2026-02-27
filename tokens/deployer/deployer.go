// | KB @CerbeRus - Nexus Invest Team
// tokens/deployer/deployer.go

package deployer

import (
	"GND/tokens/interfaces"
	"GND/tokens/registry"
	"GND/tokens/standards/gndst1"
	tokentypes "GND/tokens/types"
	coretypes "GND/types"
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Deployer отвечает за деплой токенов
type Deployer struct {
	pool         *pgxpool.Pool
	eventManager coretypes.EventManager
	evm          coretypes.EVMInterface
}

// NewDeployer создает новый экземпляр Deployer (используется в deployer_test.go и при инициализации деплоера токенов).
func NewDeployer(pool *pgxpool.Pool, eventManager coretypes.EventManager, evm coretypes.EVMInterface) *Deployer {
	return &Deployer{
		pool:         pool,
		eventManager: eventManager,
		evm:          evm,
	}
}

// DeployToken деплоит новый токен
func (d *Deployer) DeployToken(ctx context.Context, params tokentypes.TokenParams) (interfaces.TokenInterface, error) {
	// Проверяем параметры
	if params.Name == "" || params.Symbol == "" {
		return nil, errors.New("name and symbol are required")
	}
	if params.Decimals == 0 {
		params.Decimals = 18 // default decimals
	}
	if params.TotalSupply == nil || params.TotalSupply.Sign() <= 0 {
		return nil, errors.New("invalid total supply")
	}

	// Генерируем байткод
	bytecode, err := generateBytecode(params.Name, params.Symbol, params.Decimals, params.TotalSupply)
	if err != nil {
		// Отправляем событие об ошибке
		d.eventManager.Emit(&coretypes.Event{
			Type:      coretypes.EventError,
			Contract:  params.Symbol,
			Error:     fmt.Sprintf("Failed to generate bytecode: %v", err),
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("failed to generate bytecode: %v", err)
	}

	// Деплоим контракт (types.EVMInterface: from Address, gasPrice *big.Int, signature []byte)
	addr, err := d.evm.DeployContract(
		coretypes.Address(params.Owner),
		bytecode,
		coretypes.ContractMeta{
			Name:     params.Name,
			Symbol:   params.Symbol,
			Standard: params.Standard,
		},
		1000000,       // gas limit
		big.NewInt(1), // gas price
		0,             // nonce
		[]byte(nil),   // signature
		params.TotalSupply,
	)
	if err != nil {
		// Отправляем событие об ошибке
		d.eventManager.Emit(&coretypes.Event{
			Type:      coretypes.EventError,
			Contract:  params.Symbol,
			Error:     fmt.Sprintf("Failed to deploy contract: %v", err),
			Timestamp: time.Now(),
		})
		return nil, fmt.Errorf("failed to deploy contract: %v", err)
	}

	// Отправляем событие об успешном деплое
	d.eventManager.Emit(&coretypes.Event{
		Type:        coretypes.EventDeploy,
		Contract:    params.Symbol,
		FromAddress: params.Owner,
		ToAddress:   addr,
		Timestamp:   time.Now(),
		Metadata: map[string]interface{}{
			"name":         params.Name,
			"symbol":       params.Symbol,
			"decimals":     params.Decimals,
			"total_supply": params.TotalSupply.String(),
		},
	})

	// Создаем информацию о токене
	tokenInfo := tokentypes.TokenInfo{
		Address:     addr,
		Owner:       params.Owner,
		Name:        params.Name,
		Symbol:      params.Symbol,
		Decimals:    params.Decimals,
		TotalSupply: params.TotalSupply,
		Standard:    params.Standard,
		CreatedAt:   time.Now().Unix(),
	}

	// Регистрируем токен
	token, err := d.registerToken(ctx, tokenInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to register token: %v", err)
	}

	return token, nil
}

// registerToken регистрирует токен в реестре (in-memory) и при наличии pool — в БД (contracts, tokens).
func (d *Deployer) registerToken(ctx context.Context, info tokentypes.TokenInfo) (interfaces.TokenInterface, error) {
	if info.Address == "" {
		return nil, errors.New("адрес токена не задан")
	}
	totalSupply := info.TotalSupply
	if totalSupply == nil || totalSupply.Sign() <= 0 {
		totalSupply = big.NewInt(0)
	}
	token := gndst1.NewGNDst1(info.Address, info.Name, info.Symbol, info.Decimals, totalSupply, d.pool)
	token.SetInitialBalance(info.Owner, totalSupply)

	standard := info.Standard
	if standard == "" {
		standard = "GND-st1"
	}

	if err := registry.RegisterToken(info.Address, token); err != nil {
		return nil, fmt.Errorf("реестр токенов: %w", err)
	}

	if d.pool != nil {
		var contractID int
		err := d.pool.QueryRow(ctx,
			`INSERT INTO public.contracts (address, owner, created_at, type) VALUES ($1, $2, to_timestamp($3::bigint), $4) RETURNING id`,
			info.Address, info.Owner, info.CreatedAt, "token",
		).Scan(&contractID)
		if err != nil {
			return token, fmt.Errorf("запись в contracts: %w", err)
		}
		_, err = d.pool.Exec(ctx,
			`INSERT INTO public.tokens (contract_id, standard, symbol, name, decimals, total_supply) VALUES ($1, $2, $3, $4, $5, $6)`,
			contractID, standard, info.Symbol, info.Name, info.Decimals, totalSupply.String(),
		)
		if err != nil {
			return token, fmt.Errorf("запись в tokens: %w", err)
		}
	}

	return token, nil
}

// generateBytecode генерирует байткод для токена
func generateBytecode(_, _ string, _ uint8, _ *big.Int) ([]byte, error) {
	// TODO: Implement bytecode generation (name, symbol, decimals, totalSupply)
	return nil, errors.New("not implemented")
}
