# Project resume-backend-parser

Assignment about backend server for a Recruitment Management System. 

## Getting Started

About the APIs:
APIs:

1.POST /signup: Create a profile on the system (Name, Email, Password, UserType
(Admin/Applicant), Profile Headline, Address).

2. POST /login: Authenticate users and return a JWT token upon successful validation.

3. POST /uploadResume: Authenticated API for uploading resume files (only PDF or DOCX) of
the applicant. Only Applicant type users can access this API.

4. POST /admin/job: Authenticated API for creating job openings. Only Admin type users can
access this API.

5. GET /admin/job/{job_id}: Authenticated API for fetching information regarding a job
opening. Returns details about the job opening and a list of applicants. Only Admin type
users can access this API.

6. GET /admin/applicants: Authenticated API for fetching a list of all users in the system. Only
Admin type users can access this API.
7. GET /admin/applicant/{applicant_id}: Authenticated API for fetching extracted data of an
applicant. Only Admin type users can access this API.

8. GET /jobs: Authenticated API for fetching job openings. All users can access this API.

9. GET /jobs/apply?job_id={job_id}: Authenticated API for applying to a particular job. Only
Applicant users are allowed to apply for jobs

## Run in dev mode:

1. create keys for JWT
```bash
cd keys
openssl genrsa -out app.rsa 2048
openssl rsa -in app.rsa -pubout > app.rsa.pub
```

2. Setup the database
```bash
psql -U postgres -h localhost 

 \ "CREATE DATABASE resume_backend_parser"
....create user passwword grant all privilages....etc.
```

3. Make the env file:
```bash
touch .env
vim .env 
#secrets:
PORT=8080
APP_ENV=local

DB_HOST=localhost
DB_PORT=5432
DB_DATABASE=resume_backend_parser
DB_USERNAME=parser
DB_PASSWORD=
DB_SCHEMA=public
```

3. Make directories for uploads
```bash
mkdir resumes
```

2. run the application
```bash
air
```

## MakeFile

run all make commands with clean tests
```bash
make all build
```

build the application
```bash
make build
```

run the application
```bash
make run
```

Create DB container
```bash
make docker-run
```

Shutdown DB container
```bash
make docker-down
```

live reload the application
```bash
make watch
```

run the test suite
```bash
make test
```

clean up binary from the last build
```bash
make clean
```