# Coding Test
coding_test is a REST key value store service backed by etcd.

## Endpoints
The service exposes 3 endpoints for interacting with the key value store. 
| Method | URL               | Description                                                           |
| ---    | ---               | ---                                                                   |
| GET    | `/api/items/:key` | Returns the value associate with the provided key (`:key`)            |
| POST   | `/api/items/:key` | Sets the value provided in the POST body to the provided key (`:key`) |
| DELETE | `/api/items/:key` | Delete the provided key (`:key`) from the backing store               |

In all cases where a value is sent or returned from the service, it should be a JSON object with 1 string element named `"value"`.

For example to set the key `"test"` to the string value `"test value"` a POST request should be sent to the service at the URL `/api/items/test` with a request body of `{"value":"test value"}`.

## Deployment
The service was designed to be run on the single node kubernetes cluster provided by docker desktop. To run the service build the `app.dockerfile` Dockerfile as `coding-test:latest` and the `etcd.dockerfile` Docker file as `coding-test-etcd:latest` and apply the `deployment.yaml` kubernetes deployment to your kubernetes cluster. 

This service is not production ready software. Currently the service is run with a single replica for testing simplicity, to scale the service up changes would need to be made to the way etcd is deployed. 

## Security
The service as configured provides no authentification and does not perform TLS termination. Additionally if etcd is scaled up and run in a separate pod from the go application it would need to be configured with certificates as well.
