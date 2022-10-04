pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721Burnable.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

contract DeliveryToken is ERC721, ERC721Burnable {
    event NftBought(address _seller, address _buyer, uint256 _price);
    event NFTMinted(uint256 _tokenId);

    using Counters for Counters.Counter;
    Counters.Counter private _tokenIdCounter;

    // only one token can be minted from a single contract
    bool minted = false;
    // who is the current owner of the token?
    address owner;
    // only approved users can buy the token (accept delivery of shipment)
    address allowedPurchaser;
    // (contract address, tokenId) form a globally unique key
    uint256 tokenId;

    // the cost to purchase this token from the one who minted it
    uint256 public purchasePrice = 0 ether;

    string public URI = "https://example.com/metadata.json";

    constructor() ERC721("DeliveryToken", "DLV") public {
        owner = msg.sender;
    }

    // the monetaryValue is the price that it costs to buy this token
    function safeMint(address to, uint256 monetaryValue, address _allowedPurchaser) internal {
        _tokenIdCounter.increment();
        tokenId = _tokenIdCounter.current();
        minted = true;
        purchasePrice = monetaryValue;
        allowedPurchaser = _allowedPurchaser;
        _safeMint(to, tokenId);
        emit NFTMinted(tokenId);
    }

    function mintToken(uint256 monetaryValue, address _allowedPurchaser) public payable {
        require(minted == false, "A token has already been minted for this contract");
        safeMint(msg.sender, monetaryValue, _allowedPurchaser);
    }

    function getId() public view returns (uint256) {
        return tokenId;
    }

    function getOwner() public view returns (address) {
        return owner;
    }

    // Buy this token from the owner (receive the physical goods the token represents)
    // This token can only be bought and transferred once. Buyer beware!
    function buy() external payable {
        require(msg.sender == allowedPurchaser, "Not an approved user");
        require(purchasePrice == msg.value, "Wrong purchase price");

        address seller = ownerOf(tokenId);
        require(msg.sender != seller, "You cannot buy the token from yourself");

        // give the token to the buyer
        _transfer(seller, msg.sender, tokenId);
        // send ETH to the seller
        payable(seller).transfer(msg.value);
        // set the new owner
        owner = msg.sender;
        // the token cannot be repurchased
        purchasePrice = 0;

        emit NftBought(seller, msg.sender, msg.value);
    }

    function burnToken() public {
        require(msg.sender == owner, "Only the owner can void the contract");
        super._burn(tokenId);
    }
}
