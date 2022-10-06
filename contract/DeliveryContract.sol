// SPDX-License-Identifier: MIT

pragma solidity >=0.5.0;

import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721.sol";
import "../node_modules/@openzeppelin/contracts/token/ERC721/ERC721Burnable.sol";
import "../node_modules/@openzeppelin/contracts/utils/Counters.sol";

/**
 * This is a contract between a vendor and a customer, representing the agreement to deliver
 * a shipment at the agreed-upon price.
 * 
 * This contract handles the minting and transferring of ERC721 "delivery tokens". A delivery token is
 * minted by the vendor when a customer places an order. In order to receive the delivery of the
 * product, the customer must purchase the token for the sale price.
 * 
 * This contract manages the whole collection of delivery tokens that the vendor has minted.
 */
contract DeliveryContract is ERC721, ERC721Burnable {
    event NftBought(address _seller, address _buyer, uint256 _price);
    event NFTMinted(uint256 _tokenId);

    using Counters for Counters.Counter;
    Counters.Counter private _tokenIdCounter;

    // the address where money can be sent to the vendor from the customer
    address vendor;

    struct Order {
        uint256 deliveryPrice;
        uint256 orderPrice;
        address allowedRecipient;
    }

    // some mappings to keep state between the various transactions
    mapping(uint256 => Order) private orderByTokenId;
    mapping(string => uint256) private tokenIdByOrderId;
    mapping(uint256 => bool) private paidByTokenId;

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

        tokenIdByOrderId[orderId] = tokenId;
        Order memory order = Order(deliveryPrice, orderPrice, allowedPurchaser);
        orderByTokenId[tokenId] = order;
        paidByTokenId[tokenId] = false;

        // the vendor starts out owning the token because they own the goods until paid
        _safeMint(vendor, tokenId);

        // the customer is allowed to receive delivery
        approve(order.allowedRecipient, tokenId);

        emit NFTMinted(tokenId);
    }

    // This is called when the customer pays for the order. The money for the order
    // is stored in the contract, to be transferred to the vendor upon delivery.
    function payForGoods(uint256 tokenId) public payable {
        require(_exists(tokenId), "That token does not exist");
        require(paidByTokenId[tokenId] == false, "This order was paid for already");

        Order memory order = orderByTokenId[tokenId];
        require(msg.value == order.orderPrice, "Must pay for the item in full");

        paidByTokenId[tokenId] = true;
    }

    /**
     * Gets the token ID associated with an order ID
     */
    function getTokenIdForOrder(string memory orderId) public view returns (uint256) {
        return tokenIdByOrderId[orderId];
    }

    /**
     * Buy this token from the owner (receive the physical goods the token represents)
     */
    function buy(uint256 tokenId) external payable {
        require(_exists(tokenId), "That token does not exist");

        Order memory order = orderByTokenId[tokenId];
        require(msg.value == order.deliveryPrice, "Must pay for delivery");
        
        //require(address(this).balance >= totalAmount, "The vendor didn't get paid");
        require(paidByTokenId[tokenId] == true, "Order must be paid for first");

        // give the token to the buyer
        safeTransferFrom(vendor, msg.sender, tokenId);

        // send the shipping cost plus delivery cost to the seller
        address payable tokenOwner = address(uint256(vendor));
        tokenOwner.transfer(msg.value + order.orderPrice);

        emit NftBought(vendor, msg.sender, msg.value);
    }

    /**
     * Destroy the token. This equates to canceling an order that is pending shipment.
     */
    function burnTokenByOrderId(string memory orderId) public {
        uint256 tokenId = tokenIdByOrderId[orderId];
        require(tokenId != 0, "That token does not exist");
        require(ownerOf(tokenId) != vendor, "The token can only be burned after delivery");

        delete(tokenIdByOrderId[orderId]);
        delete(orderByTokenId[tokenId]);

        super._burn(tokenId);
    }

    // allows the vendor to withdraw the money from all the customer purchases and deliveries
    // function withdraw(uint256 tokenId) public {
    //     require(msg.sender == vendor, "Only the vendor can withdraw their money");
        
    //     Order memory order = orderByTokenId[tokenId];

    //     // can only withdraw money after delivery
    //     if (ownerOf(tokenId) != vendor) {
    //         payable(vendor).transfer(order.orderPrice + order.deliveryPrice);
    //     }
    // }

    /**
     * This is a hook called by the parent contract before the token is minted, transferred, or burned.
     */
    function _beforeTokenTransfer(
            address from, 
            address to, 
            uint256 tokenId) internal virtual override(ERC721) {
        
        super._beforeTokenTransfer(from, to, tokenId);

        // if the token is just being minted, it won't belong to anyone so don't check
        if (from != address(0) && to != address(0)) {
            require(_isApprovedOrOwner(to, tokenId), "not approved to transfer this token");
        }

        // if this token is being burned, 'from' is the owner and 'to' is 0
        if (to == address(0)) {
            require(_isApprovedOrOwner(from, tokenId), "not approved to burn this token");
        }
    }
}
