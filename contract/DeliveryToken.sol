pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721Burnable.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

contract DeliveryToken is ERC721, ERC721Burnable {
    event NftBought(address _seller, address _buyer, uint256 _price);
    event NFTMinted(uint256 _tokenId);

    using Counters for Counters.Counter;
    Counters.Counter private _tokenIdCounter;

    // the address where money can be sent to the vendor from the customer
    address vendor;

    mapping(uint256 => uint256) public deliveryPriceByTokenId;
    mapping(string => uint256) private tokenIdByOrderId;

    constructor() ERC721("DeliveryToken", "DLV") public {
        vendor = msg.sender;
    }

    // Minting this token represents the customer purchasing something from the vendor for delivery.
    // The customer can accept delivery by purchasing the token from the one who minted it.
    // Minting this token will transer the cost of the goods from the customer to the vendor.
    // The shipping cost is settled when the customer purchases the token.
    //
    // allowedPurchaser: the user who is allowed to purchase the token from the owner
    // deliveryCost: the price it costs the allowedPurchaser to buy this token
    function mintToken(address allowedPurchaser, uint256 deliveryPrice, uint256 orderPrice, string memory orderId) public payable {
        require(msg.sender == vendor, "Only the approved person can mint this token");
        require(msg.value == orderPrice, "Must pay for order up front");

        // Pay for the order
        payable(vendor).transfer(msg.value);

        // get a token ID and store it so we can look it up later
        _tokenIdCounter.increment();
        uint256 tokenId = _tokenIdCounter.current();
        deliveryPriceByTokenId[tokenId] = deliveryPrice;
        tokenIdByOrderId[orderId] = tokenId;

        // approve the customer to buy the token from the vendor
        approve(allowedPurchaser, tokenId);
        _safeMint(allowedPurchaser, tokenId);

        emit NFTMinted(tokenId);
    }

    function getTokenIdForOrder(string memory orderId) public view returns (uint256) {
        return tokenIdByOrderId[orderId];
    }

    function _beforeTokenTransfer(
        address from, 
        address to, 
        uint256 tokenId
    ) internal virtual override(ERC721) {
        
        super._beforeTokenTransfer(from, to, tokenId);

        require(_isApprovedOrOwner(to, tokenId));
    }


    // Buy this token from the owner (receive the physical goods the token represents)
    // This token can only be bought and transferred once. Buyer beware!
    function buy(uint256 tokenId) external payable {
        require(deliveryPriceByTokenId[tokenId] == msg.value, "Wrong purchase price");

        // give the token to the buyer
        safeTransferFrom(vendor, msg.sender, tokenId);

        // send ETH to the seller
        payable(vendor).transfer(msg.value);

        emit NftBought(vendor, msg.sender, msg.value);
    }

    function burnTokenByOrderId(string memory orderId) public {
        uint256 tokenId = tokenIdByOrderId[orderId];
        require(tokenId != 0, "That token does not exist");
        
        delete(deliveryPriceByTokenId[tokenId]);
        delete(tokenIdByOrderId[orderId]);

        super._burn(tokenId);
    }

}
