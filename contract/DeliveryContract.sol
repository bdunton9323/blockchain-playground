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
 * a promise of future delivery, assuming payment is sent. It is minted by the vendor when a customer 
 * places an order. When the order is paid for, the customer's money sits in the contract in escrow
 * until the shipment is delivered. To receive the package, the customer must purchase the token for 
 * the shipping price. The customer now owns the token and can burn it if they desire.
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
     * allowedPurchaser - the user who is allowed to purchase the token from the owner
     * deliveryPrice - the price it costs the allowedPurchaser to buy this token
     * orderPrice - the price of the goods being purchased
     * orderId - the unique ID for this order
     */
    function mintToken(
            address allowedPurchaser,
            uint256 deliveryPrice, 
            uint256 orderPrice, 
            string memory orderId) public payable {

        // each token gets a new ID
        _tokenIdCounter.increment();
        uint256 tokenId = _tokenIdCounter.current();

        // manage some internal mappings for state machine enforcement
        tokenIdByOrderId[orderId] = tokenId;
        Order memory order = Order(deliveryPrice, orderPrice, allowedPurchaser);
        orderByTokenId[tokenId] = order;
        paidByTokenId[tokenId] = false;

        // the vendor starts out owning the token because they own the goods for now
        _safeMint(vendor, tokenId);

        // the customer is allowed to receive delivery
        approve(order.allowedRecipient, tokenId);

        emit NFTMinted(tokenId);
    }

    /**
     * This is called when the customer pays for the order. The money for the order
     * is held in the contract, to be transferred to the vendor upon delivery.
     */
    function payForGoods(uint256 tokenId) public payable {
        require(_exists(tokenId), "That token does not exist");
        require(paidByTokenId[tokenId] == false, "This order was paid for already");

        Order memory order = orderByTokenId[tokenId];
        require(msg.sender == order.allowedRecipient, "Only the recipient can pay for the order");
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
     * Buy this token from the vendor (receive the physical goods the token represents)
     */
    function buy(uint256 tokenId) external payable {
        require(_exists(tokenId), "That token does not exist");
        require(paidByTokenId[tokenId] == true, "Order must be paid in full before delivery");

        Order memory order = orderByTokenId[tokenId];
        require(msg.value == order.deliveryPrice, "Must pay the shipping costs to accept delivery");
        
        // give the token to the buyer
        safeTransferFrom(vendor, msg.sender, tokenId);

        // send the shipping cost plus delivery cost to the seller
        // if the state machine is working, there is guaranteed to be enough money in the contract
        address payable tokenOwner = address(uint256(vendor));
        tokenOwner.transfer(msg.value + order.orderPrice);

        emit NftBought(vendor, msg.sender, msg.value);
    }

    /**
     * Destroy the token. Only the customer can burn the token, and only after delivery.
     */
    function burnTokenByOrderId(string memory orderId) public {
        uint256 tokenId = tokenIdByOrderId[orderId];
        require(tokenId != 0, "That token does not exist");
        require(ownerOf(tokenId) != vendor, "The token can only be burned after delivery");

        delete(tokenIdByOrderId[orderId]);
        delete(orderByTokenId[tokenId]);

        super._burn(tokenId);
    }

    /**
     * This is a hook called by the parent contract before the token is minted, transferred, or burned.
     * The base contract is more lenient than our state machine.
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
