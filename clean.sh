go clean -i -a

# clean up client build files
rm version
cd ./client
rm -rf deploy
rm -rf node_modules
