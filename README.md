# Foo bar
The goal of this project is to develop a simple REST API application that is backed by a database such as MySQL. This involves client-server environment where clients request for resources from the server managed with HTTP. The API server will include the following endpoints:
- `GET /users` : Provides the list of all the users that are logged-in
- `GET /users/me` : Allows users to access their information once they are logged in
- `PUT /users/update/me` : Updates user's info
- `GET /user/logout/me` : Logs the user out while also keeping the information in the database
- `DELETE /user/me` : Deletes the user 
- `POST /signup` : Allows users to sign up
- `POST /signin` : Authenticates the credentials and logs in
- `POST /createPost` : Creates a post
- `GET /listPost`: Lists all posts posted by all users
- `GET /readPost/:id` : Get a single post with a specific id
- `GET /post/me` : Lists all posts belonging to the current user

We are going to use JWT and auth middleware to handle all the authentication process of the users. 

In order to build a stable API, a concurrent strategy will be employed to handle large scale requests or updates using many of the built-in features in Golang. Postman or other HTTP client will be used to test the server. Once the server works, we will try to get it working locally in Docker.

## Approach ##
1. Deploy the API server in a docker container without a database, with only the `GET /users` endpoint running. The endpoint will return fixed dummy data

2. Add the database, add the `POST /signup` , the server can now create and show users

3. Add authentication and the `POST /signin` endpoint

4. Add the rest of the user endpoints involving authentication

5. Add all other endpoints including functionalities around `post`


