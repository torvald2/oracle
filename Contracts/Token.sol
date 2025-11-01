// SPDX-License-Identifier: MIT
pragma solidity ^0.8.17;

import "@openzeppelin/contracts/token/ERC721/extensions/ERC721URIStorage.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/Counters.sol";

/// @title OilField NFT (OLF)
/// @notice Simple ERC721 where each mint specifies its own metadata URL (tokenURI)
contract OilField is ERC721URIStorage, Ownable {
    using Counters for Counters.Counter;
    Counters.Counter private _tokenIds;

    constructor() ERC721("OilField", "OLF") {}

    /// @notice Mint new NFT with given metadata URL (for OpenSea, etc.)
    /// @param to address that receives the NFT
    /// @param tokenURI_ full URL to metadata (ipfs:// or https://)
    /// @return tokenId newly minted token ID
    function mint(address to, string memory tokenURI_) external onlyOwner returns (uint256) {
        _tokenIds.increment();
        uint256 newId = _tokenIds.current();
        _safeMint(to, newId);
        _setTokenURI(newId, tokenURI_);
        return newId;
    }
}
