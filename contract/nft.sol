pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
//import "../node_modules/@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

contract MyToken is ERC721 {
    event NftBought(address _seller, address _buyer, uint256 _price);
    event NFTMinted(uint256 _tokenId);

    using Counters for Counters.Counter;
    Counters.Counter private _tokenIdCounter;
    address owner;

    // the vendor is creating the token so it shouldn't cost anything to mint (aside from gas)
    uint256 public mintPrice = 0 ether;

    // the cost to purchase this token from the one who minted it
    uint256 public purchasePrice = 2 ether;

    string public URI = "https://example.com/metadata.json";

    constructor() ERC721("MyToken", "MTK") public {
        owner = msg.sender;
    }

    function safeMint(address to) internal {
        uint256 tokenId = _tokenIdCounter.current();
        _tokenIdCounter.increment();
        _safeMint(to, tokenId);
        emit NFTMinted(tokenId);
    }

    function mintToken() public payable {
        require(mintPrice == msg.value, "wrong wei sent in transaction");
        safeMint(msg.sender);
    }

    // buy this token from the owner (receive the physical goods the token represents)
    function buy(uint256 _tokenId) external payable {
        require(purchasePrice == msg.value, "Wrong purchase price");
        
        // give the token to the buyer
        address seller = ownerOf(_tokenId);
        _transfer(seller, msg.sender, _tokenId);
        
        // send ETH to the seller
        payable(seller).transfer(msg.value);

        emit NftBought(seller, msg.sender, msg.value);
    }

    // function _burn(uint256 tokenId) internal override(ERC721, ERC721URIStorage) {
    //     super._burn(tokenId);
    // }
}
