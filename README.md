# peripheral_backend-in_go

此物為把機械手臂的api寫到資料庫變成貨物並且送到交管系統

# proto generate

protoc --proto_path=./proto \
 --go_out=./protoGen --go_opt=paths=source_relative \
 --go-grpc_out=./protoGen --go-grpc_opt=paths=source_relative \
 yfy.proto

# schema

mysqldump -u root -p --no-data yfy > schema.sql
