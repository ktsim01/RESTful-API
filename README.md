# RESTful API
The goal of this project is to develop a simple REST API application. This involves client-server environment where clients can send data to the server managed with HTTP. The API server will include the following endpoints:

- `POST /signup` : User can sign up
- `POST /signin` : Users can sign in using correct credentials and returns a token that ensures authentication
- `POST /logout` : User can logout and the token is deleted
- 'GET /users : List all the users for debugging purposes
- POST /data : Sends data to the server
- GET /data : Retreives all the data send by a user
- GET /debug: Returns a string "Kyutae's awesome"

table: http verb, endpoint name, endpoint description, arguments, returns (what the endpoint sends back)

## Approach ##
1. Set up a server that can handle requests. Get it running locally. Start with the debug endpoint.
2. Add endpoints that handles users. Create a simply authentication process using tokens.
3. Add rest of the endpoints that allows users to send data.
4. Get the server running on docker.
5. Write postman tests
