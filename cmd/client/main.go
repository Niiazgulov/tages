package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"github.com/Niiazgulov/tages.git/client"
	"google.golang.org/grpc/credentials/insecure"
)

const imagePath = "../../tmp/"

func main() {
	conn, err := grpc.Dial("localhost:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot open client grpc dial connection")
	}
	client := client.NewImgWorkerClient(conn)

	files := filesInFolder()

	var actionchoice string
	fmt.Println("Введите цифру нужного действия:\n 1. Отправить новое изображение \n 2. Получить список изображений \n 3. Загрузить изображение с сервера")
	fmt.Fscan(os.Stdin, &actionchoice)

	switch actionchoice {
	case "1":
		fmt.Println("Введите номер файла:")
		for i, file := range files {
			fmt.Println(i+1, file)
		}
		var filechoice string
		fmt.Fscan(os.Stdin, &filechoice)

		filechoiceInt, err := strconv.Atoi(filechoice)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		sliceLen := len(files)
		for filechoiceInt < 0 || filechoiceInt > sliceLen {
			fmt.Printf("Введите число от 1 до %d:\n", sliceLen)
			fmt.Fscan(os.Stdin, &filechoice)
			filechoiceInt, _ = strconv.Atoi(filechoice)
		}
		selectedFile := files[filechoiceInt-1]
		client.UploadImage(strings.Join([]string{imagePath, selectedFile}, ""), selectedFile)

	case "2": // TODO: IMPLEMENT
		fmt.Println("coming soon...")

	case "3": // TODO: IMPLEMENT
		fmt.Println("coming soon...")

	default:
		fmt.Println("coming soon...")
	}

}

func filesInFolder() []string {
	dir, err := os.Open(imagePath)
	if err != nil {
		return nil
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil
	}

	filenames := []string{}
	for _, file := range files {
		filenames = append(filenames, file.Name())
	}
	return filenames
}
