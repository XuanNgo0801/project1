package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	_ "strconv"
	"strings"
	"time"

	"net/http"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) { //MessageHandler là một kiểu gọi lại có thể được thiết lập để thực thi khi có tin nhắn được xuất bản cho các chủ đề mà khách hàng đã đăng ký.
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), strings.Split(msg.Topic(), "/")[0])

	//payload cai topic insert vao ham postgres ben duoi
	deviceid := strings.Split(msg.Topic(), "/")[0]
	temperatureStr := strings.Split(string(msg.Payload()), ":")[1]
	//Convert to int64
	temperature, err := strconv.ParseInt(temperatureStr, 10, 64)
	if err != nil {
		panic(err)
	}
	time := time.Now().UnixMilli()
	//TODO get temperature value from payload. Use strings.Split :
	insertToDB(deviceid, temperature, time)
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

func main() {
	fmt.Printf("start program\n")
	go handleRequests()
	var broker = "broker.emqx.io"
	var port = 1883
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID("go_mqtt_client_2124354345") // thiết lập id của client và được sủ dụng khi kết nối với MQTT broker. 1 client id k được dài hơn 23 ký tự
	// opts.SetUsername("emqx")                              // thiết lập tên của client sử dụng khi kết nối tới MQTT
	// opts.SetPassword("public")                            // thiết lập password
	opts.SetDefaultPublishHandler(messagePubHandler) // MessageHandler sẽ được gọi khi nhận được thông báo không khớp với bất kỳ đăng ký nào đã biết. Nếu kq là true thì lệnh gọi lại không bị chặn

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts) // NewClient sẽ tạo một ứng dụng khách MQTT với tất cả các tùy chọn được chỉ định trong ClientOptions được cung cấp. Máy khách phải có phương thức Connect được gọi trên nó trước khi nó có thể được sử dụng.
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	sub(client)
	publish(client)

	client.Disconnect(250)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	//myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/temps", queryAllTemps)
	//myRouter.HandleFunc("/temp", createNewtemp).Methods("POST")
	//myRouter.HandleFunc("/temp/{id}", deletetemp).Methods("DELETE")
	//myRouter.HandleFunc("/temp/{deviceid}", returnSingletemp)
	myRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

type temps struct {
	Deviceid    string `db:"deviceid" json:"deviceid"`
	Temperature int    `db:"temperature" json:"temperature"`
	Timestamp   int    `db:"timestamp" json:"timestamp"`
}

var sensor temps

func random(sensor1 *temps) {
	rand.Seed(time.Now().UnixNano())
	min := -40
	max := 120
	sensor1.Temperature = rand.Intn(max-min+1) + min
	fmt.Println(time.Now())
}

func publish(client mqtt.Client) {
	num := 10
	for i := 0; i < num; i++ {
		fmt.Printf("publishing %d\n", i)
		if i%2 == 0 {
			random(&sensor)
			text := fmt.Sprintf("Temperature:%d", sensor.Temperature)      //sua dinh dang message
			token := client.Publish("deviceA/Temperature", 0, false, text) // them 1 so divice bat ky b c d
			token.Wait()
			time.Sleep(2 * time.Second)
		} else {
			random(&sensor)
			text := fmt.Sprintf("Temperature:%d", sensor.Temperature)      //sua dinh dang message
			token := client.Publish("deviceB/Temperature", 0, false, text) // them 1 so divice bat ky b c d
			token.Wait()
			time.Sleep(2 * time.Second)
		}
	}
}

func sub(client mqtt.Client) {
	topic := "+/Temperature"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = 123
	dbname   = "project1"
)

func insertToDB(deviceid string, temperature int64, timestamp int64) {
	db, err := sqlx.Connect("postgres", "user=postgres password=123 dbname=project1 sslmode=disable")
	if err != nil {
		fmt.Printf("error in connect db: %s", err.Error())
		log.Fatalln(err)
	}
	tx := db.MustBegin()
	tx.MustExec("INSERT INTO temperature (deviceid, temperature, timestamp) VALUES ($1, $2, $3)", deviceid, temperature, timestamp)
	tx.Commit()

}

func queryAllTemps(w http.ResponseWriter, r *http.Request) {
	result := []temps{}
	db, err := sqlx.Connect("postgres", "user=postgres password=123 dbname=project1 sslmode=disable")
	if err != nil {
		log.Fatalln(err)
	}
	// Loop through rows using only one struct
	temp := temps{}
	rows, err := db.Queryx("SELECT * FROM temperature order by deviceid, timestamp ASC")
	for rows.Next() {
		err := rows.StructScan(&temp)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("%#v\n", temp)
		result = append(result, temp)
	}

	json.NewEncoder(w).Encode(result)
}
