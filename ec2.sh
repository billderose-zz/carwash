ssh -i ./Downloads/derose.pem ubuntu@52.10.138.41
sudo apt-get install git golang
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
mkdir -p $GOPATH/src/github.com/user
go get gopkg.in/mgo.v2
go get github.com/billderose/carwash
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv 7F0CEB10
echo "deb http://repo.mongodb.org/apt/ubuntu "$(lsb_release -sc)"/mongodb-org/3.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-3.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org
sudo service mongod start
nohup sudo mongod --smallfiles &
go run server.go
crontab -l > mycron
echo "00 1 * * * mongoexport --db carwash --collection classification_labels --csv -f id,labels --out ~/labels.csv" >> mycron
crontab mycron
rm mycron
