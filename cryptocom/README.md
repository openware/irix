## Crpyto.com exchange integration
As per discussion in telegram, here's what I would like to propose.

## File Structure
  - constants.go
    - api URL (production & sandbox)
    - ws URL (production & sandbox)
    - rate limit constants
    - enums
      - Type (limit, market, stop_limit, etc)
      - Side (buy or sell)
      - etc
    - available methods (will be sorted by availability on the rest and/or ws)
      - publicInstruments
      - publicAuth
      - publicGetBook
      - privateGetWithdrawalHistory
      - etc
  - client.go
    - interface (ws & rest)
    - ws implementation
    - rest implementation
  - client_mock.go 
    - mock ws
    - mock rest
  - client_test.go
  - requests.go (contains method request definitions)
  - requests_test.go
  - types.go (contains response type struct definitions)
  - types_test.go
  - rest_actions.go (contains HTTP Rest calls)
  - rest_actions_test.go
  - ws_actions.go (contains Websocket calls)
  - ws_actions_test.go
  
## Test Structure

The current tests, imo, is not correctly testing what we need. 
It's basically just checking the response with mocked one.
In `rest_actions_test.go`, the test would verify if the method call send the right arguments to the client mock. For instance

Let's say I want to test 
```go
cryptocom.RestCreteOrder(
	instrumentName string, 
	side Side, 
	types Type, 
	price, quantity, notional number, 
	options *CreateOrderOption
)
```
with given arguments:
```go
cryptocom.RestCreateOrder("BTC_USDT", Side.buy, Type.limit, 1, 1, 1, nil)
```
the test would verify if the first argument of `rest` client mock should contains 
```go
map[string]interface{}{
    "id": 11,
    "method": "private/create-order",
    "params": {
        "instrument_name": "BTC_USDT",
        "side": "BUY",
        "type": "LIMIT",
        "price": 1,
        "quantity": 1
    }
}
```

We would also add more tests to anticipate error codes from the server based on the code response. 
We need to do this in the client test instead. 

If we also need to support `gocryptotrader.IBotExchange` interface, we will need do some tests against it.
If Tradepoint also have something similar to `IBotExchange`, we will definitely need to make test for this as well. It just needs how the interface would look like