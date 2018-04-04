go clean -i -a

# clean up client build files
rm version
cd client
rm -rf deploy
rm -rf files
rm -rf node_modules
