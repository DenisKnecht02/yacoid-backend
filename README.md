# Setup Authorizer

1. Copy file ".env.auth.sample" and rename it to ".env.auth"
2. Configure the auth environment in the file like here:

```properties
# https://docs.authorizer.dev/core/env

ENV=development # or production
DATABASE_TYPE=mongodb
DATABASE_URL=mongodb://database:27017 # database is the name of the docker container
```

3. Go into the scripts folder and execute the `start_authorizer.bat` file (docker must be running)
4. The Authorizer Dashboard is now available under <http://localhost:8080/>
5. The Authorizer Instance should now be successfully configured

<br/>
<br/>
<br/>

# Setup API

1. Copy file ".env.sample" and rename it to ".env"
2. Configure the api environment in the file like here:

```properties
DATABASE_URL=mongodb://localhost:27017
REST_PORT=3000

AUTH_CLIENT_ID=
AUTH_URL=http://localhost:8080/
AUTH_REDIRECT_URL=http://localhost:5173/
```

3. To fill `AUTH_CLIENT_ID` do the following:
   1. Navigate into `Environment > OAuth Config` in the Authorizer Dashboard
   2. Copy the client id
4. Add JWT Secrets
   1. Create a `jwtConfig.json` file in your project root folder
   2. Navigate into `Environment > JWT Secrets` in the Authorizer Dashboard
   3. Click "Copy As JSON Config"
   4. Paste the copied content into the `jwtConfig.json` file
5. Run the command `go run .` or `air` in the project root folder