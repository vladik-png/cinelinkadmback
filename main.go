package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/joho/godotenv"
)

var ec2Client *ec2.Client

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Файл .env не знайдено, використовуються системні змінні")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	ec2Client = ec2.NewFromConfig(cfg)

	http.HandleFunc("/", index)
	http.HandleFunc("/start", startInstance)
	http.HandleFunc("/stop", stopInstance)

	log.Println("AWS Admin Panel running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, r *http.Request) {
	resp, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	funcMap := template.FuncMap{
		"uppercase": strings.ToUpper,
	}

	tmpl, err := template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html")
	if err != nil {
		log.Printf("Помилка парсингу шаблону: %v", err)
		http.Error(w, "Помилка сервера", 500)
		return
	}

	err = tmpl.Execute(w, resp.Reservations)
	if err != nil {
		log.Printf("Помилка виконання шаблону: %v", err)
	}
}

func startInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	_, err := ec2Client.StartInstances(context.TODO(), &ec2.StartInstancesInput{
		InstanceIds: []string{id},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func stopInstance(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	_, err := ec2Client.StopInstances(context.TODO(), &ec2.StopInstancesInput{
		InstanceIds: []string{id},
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
