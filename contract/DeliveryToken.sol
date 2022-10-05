// SPDX-License-Identifier: MIT

pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721Burnable.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

/**
 * This contract handles the minting and transferring of "delivery tokens". A delivery token is
 * minted by the vendor when a customer places an order. In order to receive the delivery of the
 * product, the customer must purchase the token for the sale price.
 */
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

    /**
     * Minting this token represents the customer purchasing something from the vendor for 
     * delivery.
     * 
     * I could have had the customer pay now for the goods and pay the shipping later, but
     * that was more complex. Alternatively, the customer could deposit the money in the
     * contract and then on delivery it transfers to the vendor or the vendor sweeps it.
     * 
     * allowedPurchaser - the user who is allowed to purchase the token from the owner
     * deliveryPrice - the price it costs the allowedPurchaser to buy this token
     * orderPrice - the price of the goods being purchased
     */
    function mintToken(
            address allowedPurchaser, 
            uint256 deliveryPrice, 
            uint256 orderPrice, 
            string memory orderId) public payable {

        // each token gets a new ID
        _tokenIdCounter.increment();
        uint256 tokenId = _tokenIdCounter.current();

        deliveryPriceByTokenId[tokenId] = orderPrice + deliveryPrice;
        tokenIdByOrderId[orderId] = tokenId;

        // the vendor starts out owning the token because they own the goods until paid
        _safeMint(vendor, tokenId);

        // approve the customer to buy the token from the vendor
        approve(allowedPurchaser, tokenId);

        emit NFTMinted(tokenId);
    }

    /**
     * Gets the token ID associated with an order ID
     */
    function getTokenIdForOrder(string memory orderId) public view returns (uint256) {
        return tokenIdByOrderId[orderId];
    }

    /**
     * This is a hook called by the parent contract before the token is minted, transferred, or burned.
     */
    function _beforeTokenTransfer(
            address from, 
            address to, 
            uint256 tokenId) internal virtual override(ERC721) {
        
        super._beforeTokenTransfer(from, to, tokenId);

        // if the token is just being minted, it won't belong to anyone so don't check
        if (from != address(0)) {
            require(_isApprovedOrOwner(to, tokenId));
        }
    }

    /**
     * Buy this token from the owner (receive the physical goods the token represents)
     */
    function buy(uint256 tokenId) external payable {
        require(deliveryPriceByTokenId[tokenId] == msg.value, "Wrong purchase price");

        // give the token to the buyer
        safeTransferFrom(vendor, msg.sender, tokenId);

        // send ETH to the seller
        address payable tokenOwner = address(uint256(vendor));
        tokenOwner.transfer(msg.value);

        emit NftBought(vendor, msg.sender, msg.value);
    }

    /**
     * Destroy the token. This equates to canceling an order that is pending shipment.
     */
    function burnTokenByOrderId(string memory orderId) public {
        uint256 tokenId = tokenIdByOrderId[orderId];
        require(tokenId != 0, "That token does not exist");
        
        delete(deliveryPriceByTokenId[tokenId]);
        delete(tokenIdByOrderId[orderId]);

        super._burn(tokenId);
    }
}
