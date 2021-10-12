TARGET_DIR=("ealiyun" "edingtalk" "ees" "eetcd" "egitlab" "egorm" "ehuawei" "ejenkins" "ejira" "ek8s" "ekafka" "elogger" "eminio" "emns" "emongo" "eoauth2" "eredis" "esession" "etoken" "ewechat")
ROOT=$(pwd)
for dir in ${TARGET_DIR[@]}; do
    cd $dir && go build -v ./...
    cd $ROOT
done