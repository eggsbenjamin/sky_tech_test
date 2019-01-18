## Tech Test

## Task

Create an order service that interacts with a pre-existing order process service.

The order process service takes care of managing the lifecycle of an order process e.g. handling payments, inventory, shipping etc...

The order service should allow provide functionality for creating and querying orders.

An order can be in one of the following states:
  - CREATED
  - PROCESSING
  - FULFILLED
  - FAILED

When an order is first created should have the status CREATED.

When the order is processing (i.e an order process belonging to the order has been created) it should have the status PROCESSING.

If the order process has the status SUCCEEDED then the order should have the status FULFILLED.

If the order process has the status FAILED then the order should have the status FAILED.

The order service should expose the following end points:

#### Endpoints

`GET /order`

Returns all orders that meet the supplied filter(s). If no filter parameters are supplied then all orders should be returned.

##### Parameters

`status` (optional) - filters orders by status.

`GET /order/{order_id}`

Returns an order by id.

`POST /order`

Creates an order.

##### Body

`id` (mandatory) - the id of the order to be created.

### Order Process Service

RESTful order process service that provides order processing execution and query functionality. 

An order is an entity that an order process belongs to. 
An order process references an order by the order's id which is supplied when creating an order process.
An order may have zero or one processes belonging to it.

An order process can be in one of the following statuses:
  - RUNNING
  - SUCCEEDED
  - FAILED

### Setup

`docker run -p 8000:8000 eggsbenjamin/order_process_service`

#### Endpoints

`GET /order_process`

Returns all order processes or a single order process if an order id is supplied in the querystring.

##### Parameters

`order_id` (optional) - filters the order processes by order id.

`POST /order_process`

Creates and executes an order process.

##### Body

`order_id` (mandatory) - the id of the order that the order process belongs to.

##### Parameters

`callback_url` (optional) - a url to post the result of the process to on completion. Must be fully qualified including scheme!
