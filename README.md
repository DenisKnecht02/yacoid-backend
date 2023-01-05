# Workflow for Development
## Setup Authorizer

1. Copy file `.env.auth.sample` and rename it to `.env.auth`
2. Configure the auth environment in the file like here:

```properties
# https://docs.authorizer.dev/core/env

ADMIN_SECRET=
ENV=development
DATABASE_TYPE=mongodb
DATABASE_URL=mongodb://database:27017 # database is the name of the docker container
ROLES=user
PROTECTED_ROLES=moderator,admin
```

3. Go into the scripts folder and execute the `start_authorizer.bat` script (docker must be running)
4. The Authorizer Dashboard is now available under <http://localhost:8080/>
5. When logging in enter the password specified by `ADMIN_SECRET`
6. The Authorizer Instance should now be successfully configured and can be accessed

<br/>

## Setup API

1. Copy file `.env.sample` and rename it to `.env`
2. Configure the api environment in the file like here:

```properties
DATABASE_URL=mongodb://localhost:27017
REST_PORT=3000

AUTH_CLIENT_ID=
AUTH_ADMIN_SECRET=
AUTH_URL=http://localhost:8080
AUTH_REDIRECT_URL=http://127.0.0.1:5173
```

3. To fill `AUTH_CLIENT_ID` do the following:
   1. Navigate into `Environment > OAuth Config` in the Authorizer Dashboard
   2. Copy the client id
   
4. `AUTH_ADMIN_SECRET` needs to be the same as `ADMIN_SECRET` from the `.env.auth` file

## Start the system
1. Run the command `go run .` or `air` in the project root folder
2. The individual API endpoints can be accessed under <http://localhost:3000/>

<br/>
<br/>

# Workflow for Production

## Setup Authorizer
1. Copy file `.env.auth.sample` and rename it to `.env.prod.auth`
2. Configure the auth environment in the file like here:

```properties
# https://docs.authorizer.dev/core/env

ADMIN_SECRET=
ENV=production
DATABASE_TYPE=mongodb
DATABASE_URL=mongodb://database:27017 # database is the name of the docker container
ROLES=user
PROTECTED_ROLES=moderator,admin
```
<br/>

## Setup API

1. Copy file `.env.sample` and rename it to `.env.prod`
2. Configure the api environment in the file like here:

```properties
DATABASE_URL="mongodb://database:27017"
REST_PORT=3000

AUTH_CLIENT_ID=
AUTH_ADMIN_SECRET=
AUTH_URL=http://authorizer:8080
AUTH_REDIRECT_URL=http://127.0.0.1:5173
```

3. If you already successfully started Authorizer in development, then you can fill the existing `AUTH_CLIENT_ID`. Otherwise this will be done in the next steps.
4. `AUTH_ADMIN_SECRET` needs to be the same as `ADMIN_SECRET` from the `.env.prod.auth` file

<br/>

## Start the system
1. Go into the scripts folder and execute the `start_production_system.bat` script (docker must be running)
2. If you already filled `AUTH_CLIENT_ID`:
   1. The system is now successfully running
   2. The Authorizer Dashboard is available under <http://localhost:8080/>
      1. When logging in enter the password specified by `ADMIN_SECRET` in the `.env.prod.auth` or `.env.prod` file
   4. The API endpoints are available under <http://localhost:3000/>
3. If you didn't fill `AUTH_CLIENT_ID`:
   1. Then the first time the Docker container `yacoid-api-container` will fail to start, because the `AUTH_CLIENT_ID` is missing in the `.env.prod` file.
   2. To fill `AUTH_CLIENT_ID` do the following:
      1. The Authorizer Dashboard is still available under <http://localhost:8080/>
      2. Navigate into `Environment > OAuth Config` in the Authorizer Dashboard
      3. Copy the client id
   3. Now terminate the current process and repeat the current section `Start the system`