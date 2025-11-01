// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import {FunctionsClient} from "@chainlink/contracts@1.5.0/src/v0.8/functions/v1_0_0/FunctionsClient.sol";
import {FunctionsRequest} from "@chainlink/contracts@1.5.0/src/v0.8/functions/v1_0_0/libraries/FunctionsRequest.sol";
import {ConfirmedOwner} from "@chainlink/contracts@1.5.0/src/v0.8/shared/access/ConfirmedOwner.sol";

/**
 * @title ValuationFunctionsConsumer
 * @notice Chainlink Functions consumer that fetches oil well valuation data from your API
 * @dev Designed for testing on Sepolia
 */
contract ValuationFunctionsConsumer is FunctionsClient, ConfirmedOwner {
  using FunctionsRequest for FunctionsRequest.Request;

  bytes32 public s_lastRequestId;
  bytes public s_lastResponse;
  bytes public s_lastError;

  error UnexpectedRequestID(bytes32 requestId);

  event Response(bytes32 indexed requestId, string result, bytes response, bytes err);

  // Router for Sepolia
  address router = 0xb83E47C2bC239B3bf370bc41e1459A34b41238D0;

  // donID for Sepolia
  bytes32 donID = 0x66756e2d657468657265756d2d7365706f6c69612d3100000000000000000000;

  // Gas limit for fulfill
  uint32 gasLimit = 300_000;

  // The latest parsed valuation data (encoded as a string)
  string public valuationData;

  /**
   * @notice JS source code that runs in the Chainlink Functions environment
   * @dev It fetches the valuation data and returns compacted string of key fields
   */
  string source = string.concat(
    "const id = args[0];",
    "const apiResponse = await Functions.makeHttpRequest({",
      "url: `https://outfromit.com/valuation/${id}`,",
    "});",
    "if (apiResponse.error) { throw Error('Request failed'); }",
    "const data = apiResponse.data;",
    "const result = [",
      "data.NpvUsd,",
      "data.MarketValueUsd,",
      "data.DiscountPct ?? 'null',",
      "data.Confidence,",
      "data.RemainingReservesBbl,",
      "data.OilPriceUsd,",
      "data.OperatingCostPerBbl,",
      "data.DiscountRate,",
      "data.RoyaltyRate",
    "].join('|');",
    "return Functions.encodeString(result);"
  );

  constructor() FunctionsClient(router) ConfirmedOwner(msg.sender) {}

  /**
   * @notice Sends the Chainlink Functions request
   * @param subscriptionId LINK subscription ID
   * @param args Arguments for the request (the first one should be the valuation ID)
   */
  function sendRequest(
    uint64 subscriptionId,
    string[] calldata args
  ) external onlyOwner returns (bytes32 requestId) {
    FunctionsRequest.Request memory req;
    req.initializeRequestForInlineJavaScript(source);
    if (args.length > 0) req.setArgs(args);

    s_lastRequestId = _sendRequest(req.encodeCBOR(), subscriptionId, gasLimit, donID);
    return s_lastRequestId;
  }

  /**
   * @notice Callback called by Chainlink upon fulfillment
   */
  function fulfillRequest(
    bytes32 requestId,
    bytes memory response,
    bytes memory err
  ) internal override {
    if (s_lastRequestId != requestId) {
      revert UnexpectedRequestID(requestId);
    }
    s_lastResponse = response;
    valuationData = string(response);
    s_lastError = err;

    emit Response(requestId, valuationData, response, err);
  }
}
