# RESTful API
The goal of this project is to develop a simple REST API application. This involves client-server environment where clients can send data to the server managed with HTTP. Endpoint included in the API server is represented in the following table:

| HTTP Verb | Endpoint Name | Endpoint Description | Arguments | Return Value |
|     ---      |      ---       |      ---      | --- | --- |
| POST  | /signup     | User can sign up    | None | Success/ Error|
| POST  | /signin   |  Users can sign in using correct credentials and returns a token that ensures authentication   | User ID, Password| Authentication token|
| GET   | /welcome | Welcomes valid users and sessions | Session Token | Welcome/ Error message|
| POST  | /logout | User can logout and the token is deleted | None | Success/ Error message|
| Post  | /refresh | Users can refresh their session token | None | Succes/ Error message| 
| GET   | /users | List all the users for debugging purposes | None | List of users |
| POST | /data | Sends data to the server | Data | Success/ Error message |
| GET | /data | Retreives all the data sent by a user | User ID | Success/ Error message |
| GET | /debug | Returns a string "Kyutae's awesome" | None | String |

## Approach ##
1. Set up a server that can handle requests. Get it running locally. Start with the debug endpoint.
2. Add endpoints that handles users. Create a simply authentication process using tokens.
3. Add rest of the endpoints that allows users to send data.
4. Get the server running on docker.
5. Write postman tests
