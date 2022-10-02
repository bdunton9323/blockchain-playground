pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

contract MyToken is ERC721 {
    event NftBought(address _seller, address _buyer, uint256 _price);
    event NFTMinted(uint256 _tokenId);

    using Counters for Counters.Counter;
    Counters.Counter private _tokenIdCounter;
    address owner;

    // (contract address, tokenId) form a globally unique key
    uint256 tokenId;

    // the cost to purchase this token from the one who minted it
    uint256 public purchasePrice = 0 ether;

    string public URI = "https://example.com/metadata.json";

    constructor() ERC721("MyToken", "MTK") public {
        owner = msg.sender;
    }

    function safeMint(address to, uint256 monetaryValue) internal {
        tokenId = _tokenIdCounter.current();
        purchasePrice = monetaryValue;
        _tokenIdCounter.increment();
        _safeMint(to, tokenId);
        emit NFTMinted(tokenId);
    }

    function mintToken(uint256 monetaryValue) public payable {
        require(monetaryValue != 0, "Must have a nonzero value");
        require(purchasePrice == 0, "A token has already been minted");
        safeMint(msg.sender, monetaryValue);
    }

    function getId() public view returns (uint256) {
        return tokenId;
    }

    // buy this token from the owner (receive the physical goods the token represents)
    function buy(uint256 _tokenId) external payable {
        require(purchasePrice == msg.value, "Wrong purchase price");
        require(purchasePrice != 0, "The token can only be bought once");

        address seller = ownerOf(_tokenId);
        require(msg.sender != seller, "You cannot buy the token from yourself");

        // give the token to the buyer
        _transfer(seller, msg.sender, _tokenId);
        // send ETH to the seller
        payable(seller).transfer(msg.value);
        // set the new owner
        owner = msg.sender;
        // the token cannot be repurchased
        purchasePrice = 0;

        emit NftBought(seller, msg.sender, msg.value);
    }

    // function _burn(uint256 tokenId) internal override(ERC721, ERC721URIStorage) {
    //     super._burn(tokenId);
    // }
}
