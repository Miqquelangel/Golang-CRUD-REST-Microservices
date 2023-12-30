# Microservices

The purpose of this project is to create a CRUD API in Golang, whose source of data will be a MongoDB deployed through minikube. Once the API is created, it will be containerized and deployed into our minikube setup, providing us a basic CRUD microservice setup.

In order to automate future releases of the API and to customize our setup, the microservices will be deployed through ArgoCD, using Kustomize. This project will not be focused on how to configure ArgoCD and Kustomize, however, the source code will be available under the 'argocd’ directory.

The API will provide 4 endpoints (“/login”, “/signup”, “/auth” and /operate), which functions goes as follows:

1.- “/login” endpoint, will offer POST and OPTIONS requests. The POST requests will generate a JWT token (HTTP 200) that will be used as a source of truth to authenticate users into our “/auth” endpoint.

2.- “/signup” endpoint, will offer POST and OPTIONS requests as the “/login” endpoint. The main difference is that the POST requests will insert the users data into the mongoDB (HTTP 201), as well as generate a JWT token for authentication to the “/auth” endpoint.

3.- “/auth” endpoint, will offer POST and OPTIONS requests. The main function of this endpoint is to fetch if the JWT token is valid. If it is (HTTP 200), it will return a custom message like "Hello! User ($username) has been authenticated". If not, it will return a HTTP 401 unauthorized.

As for the mongoDB database, only two databases with the following collections will be created:

1.- DB “users”, collection “data”. The collection “data”, will store all the fields that are relevant to the user, in this case will be -> ‘username’ and ‘password’.

2.- DB “logs”, collections “login”, “signup” and “auth”. Each collection will store information of every request to their corresponding endpoint. The information gathered will be as follows -> requestURI, httpStatus, source IP, Date, Header and Host.
