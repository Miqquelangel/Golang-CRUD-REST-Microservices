# Microservices

The purpose of this project is to create a CRUD API in Golang, whose source of data will be a MongoDB deployed through minikube. Once the API is created, it will be containerized and deployed into our minikube setup, providing us a basic CRUD microservice setup.

The API will provide 4 endpoints (“/login”, “/signup”, “/auth” and /operate), which functions goes as follows:

1.- “/login” endpoint, will offer POST and OPTIONS requests. The POST requests will generate a JWT token (HTTP 200) that will be used as a source of truth to authenticate users into our “/auth” endpoint.

2.- “/signup” endpoint, will offer POST and OPTIONS requests as the “/login” endpoint. The main difference is, as obvious, that the POST requests will insert the users data into the mongoDB (HTTP 201) as well as generate a JWT token for authentication to the “/auth” endpoint.

3.- “/auth” endpoint, will offer POST and OPTIONS requests. The main function of this endpoint is to fetch if the JWT token is valid. If it is (HTTP 200), it will return a custom message like “Hello. $user authenticated”. If not, it will return a HTTP 401 unauthorized.

4.- “/operate” endpoint, will offer GET, POST, PUT, DELETE and OPTIONS requests. This endpoint will be used to manipulate data from our database at will.

As for the mongoDB database, only two databases with the following collections will be created:

1.- DB “users”, collection “data”. The collection “data”, will store all the fields that are relevant to the user, in this case will be -> ‘username’, ‘email’, ‘password’ and ‘created_at’

2.- DB “logs”, collections “login”, “signup” and “auth”. Each collection will store information of every request to their corresponding endpoint. The information gathered will be as follows:
