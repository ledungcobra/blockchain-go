# blockchain-go

#Cách chạy app:
##Bước 1:
	> tạo thư mục db trong folder source code để chứa cơ sở dữ liệu block chain của từng node khác nhau
	>mỗi node phân biệt với nhau bằng biến môi trường NODE_ID

##Bước 2:
	>chạy lệnh $env:NODE_ID=3000 tiếp theo chạy lệnh go run main.go init (để init block chain)
	>sau khi chạy ta sẽ nhận được 3 file wallet1.json, wallet2.json, wallet3.json là 3 địa chỉ ví 
	>đồng thời wallet1.json sẽ nhận dc 100 tiền khởi tạo blockchain
> sau đó chạy lệnh go run main.go startnode
> tiếp theo chạy lệnh ở một terminal khác để chạy miner $env:NODE_ID=4000 && go run main.go startnode -mine <ĐỊA CHỈ MINER> (lấy địa chỉ miner trong wallet.json)
##Bước 3:
	> Đổi tên wallet_3000.dat trong thư mục source code vừa được khởi tạo thành wallet_5000 để chạy web wallet
##Bước 4:
	>Mở terminal mới chạy lệnh $env:NODE_ID=5000 sau đó chạy lệnh go run main.go runweb để chạy web wallet tại port 8080 và truy cập vào wallet 

##Bước 5:
	>Tại thư mục source web react chạy lệnh npm start để bật web wallet

Tham khảo source code tại [Github Blockchain Go](https://github.com/Jeiwan/blockchain_go)

Nguồn tài liệu đọc: [Jeiwan Medium](https://jeiwan.medium.com)

Video demo tại: [Block chain demo youtube](https://youtu.be/Ehry4Khx6wc)