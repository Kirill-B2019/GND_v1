package integration

import (
	"bytes"
	"context"
	_ "errors"
	"io"
	"time"

	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSClient - обёртка для взаимодействия с IPFS-нодой
type IPFSClient struct {
	sh *shell.Shell
}

// NewIPFSClient создает клиента для работы с IPFS по адресу ноды (например, "localhost:5001")
func NewIPFSClient(apiAddr string) *IPFSClient {
	return &IPFSClient{
		sh: shell.NewShell(apiAddr),
	}
}

// AddData загружает данные (например, метаданные или файл) в IPFS и возвращает CID
func (c *IPFSClient) AddData(data []byte) (string, error) {
	reader := bytes.NewReader(data)
	cid, err := c.sh.Add(reader)
	if err != nil {
		return "", err
	}
	return cid, nil
}

// AddFile загружает файл в IPFS и возвращает CID
func (c *IPFSClient) AddFile(r io.Reader) (string, error) {
	cid, err := c.sh.Add(r)
	if err != nil {
		return "", err
	}
	return cid, nil
}

// GetData скачивает данные по CID из IPFS
func (c *IPFSClient) GetData(cid string) ([]byte, error) {
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	reader, err := c.sh.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// PinCID закрепляет (pin) объект по CID в локальной ноде IPFS
func (c *IPFSClient) PinCID(cid string) error {
	return c.sh.Pin(cid)
}

// UnpinCID удаляет pin по CID
func (c *IPFSClient) UnpinCID(cid string) error {
	return c.sh.Unpin(cid)
}

// CheckCIDExists проверяет, доступен ли объект по CID
func (c *IPFSClient) CheckCIDExists(cid string) (bool, error) {
	reader, err := c.sh.Cat(cid)
	if err != nil {
		return false, err
	}
	defer reader.Close()
	buf := make([]byte, 1)
	_, err = reader.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	return true, nil
}

// Пример интеграции с контрактами/метаданными:
// 1. Загружаем метаданные в IPFS, сохраняем CID в блокчейне.
// 2. По CID любой участник может получить файл/метаданные через IPFS.
// Инициализация клиента (локальный IPFS-демон)
/*
Как использовать
ipfs := integration.NewIPFSClient("localhost:5001")

// Загрузка данных
cid, err := ipfs.AddData([]byte("hello, Ganymede!"))
if err != nil {
// обработка ошибки
}

// Получение данных по CID
data, err := ipfs.GetData(cid)*/
