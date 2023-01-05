docker pause yacoid-auth-container
docker pause yacoid-mongodb-container
docker stop yacoid-auth-container
docker stop yacoid-mongodb-container
docker rm yacoid-auth-container
docker rm yacoid-mongodb-container
docker rmi lakhansamani/authorizer