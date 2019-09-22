# Chirper

Chirper is an example REST API server prepared to scale well in the public cloud infrastructure. It's written in Go 1.13.
The application does not have many crucial functionalities implemented, such as proper authentication and authorization, user management, etc.

The Chirper was created as a prototype to present the understanding of the Go language and its conventions.

## Setup
You can either use a local development environment or a dockerized development environment.

### Dockerized development environment
1. Install Docker.
2. Install docker-compose that supports configuration files with version 3.
3. `make up`

The docker-compose.yml should setup the standalone version of the server along with two databases - one for development and one for testing.
Data persistence was not yet implemented so once you delete the stack with `make down` all the data from the database will be lost.


### Local development environment
1. Install Go programming language (version 1.13).
2. Install PostgreSQL (version 11.5).
3. Create a database and a role with proper access.
4. `cp cmd/standalone/.env.example cmd/standalone/.env`
5. Fill the .env with PostgreSQL connection data.
6. `go run cmd/standalone/main.go`


## Tests
As time didn't allow to provide the full test process, only the most important integration tests were implemented.

They can be run with `make test` with dockerized development environment.

I implemented the tests using a real database connection instead of mocks. This approach allows me to test the data that's processed and not only the SQL queries. Testing with real databases is recommended by some known people in the Golang community, eg. Jon Calhoun.
It helps to deal with libraries that chain methods:
https://www.calhoun.io/go-experience-report-interfaces-with-methods-that-return-themselves/

It also helps with keeping the tests up to date after big refactors of the queries.

# Shortcuts
1. Tests
It's obvious that the project lacks more tests. I focused on other parts of the task and created only the most important test scenarios.
Aside from modules that are not yet tested at all, I'd add more test cases to the modules that contain tests. They only check one positive scenario for each function. It's important to also check failing scenarios.
Golang also provides a good way to test HTTP handlers.

2. Database Queries
Some of the queries can be written better. But most importantly, in the future we should group multiple queries in batches to avoid overscaling the database.

3. Authentication/Authorization
I only mocked the middleware responsible for Authentication/Authorization. If I had more time I'd implement JWT tokens.
I'd hash the user passwords with bcrypt: https://github.com/golang/crypto/blob/master/bcrypt/bcrypt.go

4. Pagination
Now there's no sophisticated way to present big amounts of data to the user. The pagination is required for any API servers.

5. TLS
The services would run in a private network, but still need a certificate to connect securely with the end-user.

6. Config abstraction
In a real production-ready application there should be some kind of abstraction over loading configs to the application. The code should access variables by a proper interface, not through the raw os.Getenv.

7. Makefile
Makefile should be the main interface for controlling development behavior of the application. It should be refactored, expanded and commented properly.

8. Infrastructure as Code
The AWS infrastructure should be built using CloudFormation, Terraform or other IaC language. It should not be configured by hand. Not a single resource. The IaC template serves as good documentation of the resources.

9. CI/CD
There are many CI/CD providers that integrate with AWS. The AWS even has its own pipeline. For the production deployment, I'd set up the pipeline that best suits the code repository and dependencies.

# Production Deployment

For the application that needs to scale for millions of users, the infrastructure environment is important as much as the code.

I would deploy the application to AWS ECS. As you can see in the project repository, the `cmd` directory contains applications separated by the endpoints (`cmd/services/*`). They can be deployed independently and set up with their own autoscaling groups.
This way we can scale up the GET endpoints that are going to take most of the traffic.

I would also set up a load balancer to split the traffic evenly between nodes.

AWS provides an excellent approach to application security. In the production environment, all secrets can be stored securely in the SSM Parameter Store. All access can be set up with roles instead of key pairs. The secrets can be downloaded and injected into environment variables in the runtime.

The Dockerfiles for ECS should be different that the ones used for development. It's important to minimize the size of the images by using multi-stage building. Example Dockerfiles are available in the `cmd/services/*` subdirectories.
