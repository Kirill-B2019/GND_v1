const SimpleGNDToken = artifacts.require("SimpleGNDToken");

contract("SimpleGNDToken", accounts => {
    it("should mint initial supply to owner", async () => {
        const instance = await SimpleGNDToken.new(1000);
        const balance = await instance.balanceOf(accounts[0]);
        assert.equal(balance.toString(), "1000");
    });
});
