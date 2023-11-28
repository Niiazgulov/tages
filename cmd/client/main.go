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

const (
	imagePath = "../../tmp/"
)

func main() {
	conn, err := grpc.Dial("localhost:44044", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("cannot open client grpc dial connection")
	}
	client := client.NewImgWorkerClient(conn)

	files := filesInFolder()

	var actionchoice string
	fmt.Println("Введите цифру нужного действия:\n 1. Отправить новое изображение на жесткий диск \n 2. Просмотр списка всех загруженных файлов на жестком диске \n 3. Загрузить изображение с сервера")
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

	case "2":
		res, err := client.InformImage()
		if err != nil {
			log.Fatal("cannot get all images info into client main")
		}
		sliceofslice := res.GetResponse()
		fmt.Printf("|%15s|%35s|%35s|\n", "Имя файла", "Дата создания", "Дата обновления")
		for _, v := range sliceofslice {
			slice := v.GetValue()
			fmt.Printf("|%15s|%35s|%35s|\n", slice[0], slice[1], slice[2])
		}

	case "3":
		fmt.Println("Введите номер файла, который вы хотите получить:")
		fmt.Printf("|%15s|%15s|%35s|%35s|\n", "Номер файла", "Имя файла", "Дата создания", "Дата обновления")

		res, err := client.InformImage()
		if err != nil {
			log.Fatal("cannot get all images info into client main")
		}
		sliceofslice := res.GetResponse()

		for i, v := range sliceofslice {
			slice := v.GetValue()
			fmt.Printf("|%15d|%15s|%35s|%35s|\n", i+1, slice[0], slice[1], slice[2])
		}

		var filechoiceInt int
		fmt.Fscan(os.Stdin, &filechoiceInt)
		filechoice := files[filechoiceInt-1]

		res2, err := client.DownloadImage(filechoice)
		if err != nil {
			log.Fatal("cannot get image into client main")
		}

		imgByte := res2.ImageData
		err = os.WriteFile(filechoice, imgByte, 0644)
		if err != nil {
			log.Fatal("cannot save byte array into file (client main)")
		}

		log.Printf("Файл %s успешно сохранен!", filechoice)

	default:
		fmt.Println("Попробуйте еще раз. Нужно ввести цифру от 1 до 3")
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
