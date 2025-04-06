package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/quotedprintable"
	"os"
	"path/filepath"
	"strings"
)

type Contact struct {
	Index  int
	FN     string              // Полное имя
	N      []string            // Структурированное имя: Фамилия, Имя, Отчество и т.д.
	TEL    []string            // Номера телефонов
	EMAIL  []string            // Email адреса
	ORG    []string            // Организации
	TITLE  string              // Должность
	ADR    []string            // Адреса
	URL    []string            // Веб-сайты
	NOTE   []string            // Заметки
	PHOTO  []byte              // Фото (base64 или ссылка)
	Fields map[string][]string // Прочие поля
}

// decodeQPString декодирует quoted-printable строку в UTF-8
func decodeQPString(s string) string {
	r := quotedprintable.NewReader(strings.NewReader(s))
	decoded, err := io.ReadAll(r)
	if err != nil {
		return s // если ошибка, вернуть оригинал
	}
	return string(decoded)
}

// readVCFCards читает файл и возвращает срез блоков vCard (между BEGIN:VCARD и END:VCARD)
func readVCFCards(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cards []string
	var current []string
	inCard := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "BEGIN:VCARD" {
			inCard = true
			current = []string{line}
			continue
		}
		if inCard {
			current = append(current, line)
			if strings.TrimSpace(line) == "END:VCARD" {
				cards = append(cards, strings.Join(current, "\n"))
				inCard = false
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

// parseContacts разбирает каждый блок vCard и формирует слайс структур Contact
func parseContacts(blocks []string) []Contact {
	var contacts []Contact
	index := 1

	for _, block := range blocks {
		c := Contact{
			Index:  index,
			Fields: make(map[string][]string),
		}

		lines := strings.Split(block, "\n")

		keys := []string{"FN", "N", "TEL", "EMAIL", "ORG", "TITLE", "ADR", "URL", "NOTE", "PHOTO", "END"}
		var cleaned []string

		for i := 0; i < len(lines); i++ {
			line := lines[i]
			isContinuation := true
			for _, k := range keys {
				if strings.HasPrefix(strings.ToUpper(line), k) {
					isContinuation = false
					break
				}
			}
			if isContinuation && len(cleaned) > 0 {
				prev := cleaned[len(cleaned)-1]
				if strings.HasSuffix(prev, "=") {
					prev = strings.TrimSuffix(prev, "=")
				}
				cleaned[len(cleaned)-1] = prev + strings.TrimSpace(line)
			} else {
				cleaned = append(cleaned, line)
			}
		}
		lines = cleaned

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "BEGIN") || strings.HasPrefix(line, "END") {
				continue
			}

			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			keyVal := strings.ToUpper(parts[0])
			value := strings.TrimSpace(parts[1])

			switch {
			case strings.HasPrefix(keyVal, "FN"):
				c.FN = decodeQPString(value)
			case strings.HasPrefix(keyVal, "N"):
				decoded := decodeQPString(value)
				c.N = strings.Split(decoded, ";")
			case strings.HasPrefix(keyVal, "TEL"):
				c.TEL = append(c.TEL, value)
			case strings.HasPrefix(keyVal, "EMAIL"):
				c.EMAIL = append(c.EMAIL, value)
			case strings.HasPrefix(keyVal, "ORG"):
				c.ORG = append(c.ORG, decodeQPString(value))
			case strings.HasPrefix(keyVal, "TITLE"):
				c.TITLE = decodeQPString(value)
			case strings.HasPrefix(keyVal, "ADR"):
				c.ADR = append(c.ADR, decodeQPString(value))
			case strings.HasPrefix(keyVal, "URL"):
				c.URL = append(c.URL, value)
			case strings.HasPrefix(keyVal, "NOTE"):
				c.NOTE = append(c.NOTE, decodeQPString(value))
			case strings.HasPrefix(keyVal, "PHOTO"):
				data := strings.ReplaceAll(value, "\n", "")
				data = strings.ReplaceAll(data, "\r", "")
				data = strings.ReplaceAll(data, " ", "")
				decoded, err := base64.StdEncoding.DecodeString(data)
				if err == nil {
					c.PHOTO = decoded
				}
			default:
				c.Fields[keyVal] = append(c.Fields[keyVal], value)
			}
		}

		contacts = append(contacts, c)
		index++
	}

	return contacts
}

// printContact выводит структуру Contact в человекочитаемом виде
func printContact(c Contact) {
	fmt.Printf("Контакт #%d\n", c.Index)
	fmt.Println("Полное имя:", c.FN)
	fmt.Println("Структурированное имя:", c.N)
	fmt.Println("Телефоны:", c.TEL)
	fmt.Println("Email:", c.EMAIL)
	fmt.Println("Организация:", c.ORG)
	fmt.Println("Должность:", c.TITLE)
	fmt.Println("Адрес:", c.ADR)
	fmt.Println("Веб-сайты:", c.URL)
	fmt.Println("Заметки:", c.NOTE)
	fmt.Println("Фото:", func() string {
		if len(c.PHOTO) > 0 {
			return fmt.Sprintf("[base64] %d байт", len(c.PHOTO))
		}
		return "—"
	}())
	if len(c.Fields) > 0 {
		fmt.Println("Дополнительные поля:")
		for k, v := range c.Fields {
			fmt.Printf("  %s: %s\n", k, strings.Join(v, ", "))
		}
	}
	fmt.Println(strings.Repeat("-", 40))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("❗ Использование: go run main.go <путь-к-vcf>")
		return
	}

	vcfPath := os.Args[1]

	contacts, err := readVCFCards(vcfPath)
	if err != nil {
		log.Fatal("Ошибка чтения VCF:", err)
	}
	fmt.Println("✅ Всего контактов:", len(contacts))

	contactList := parseContacts(contacts)

	outputHTML := strings.TrimSuffix(filepath.Base(vcfPath), filepath.Ext(vcfPath)) + "_report.html"
	//outputPDF := strings.TrimSuffix(filepath.Base(vcfPath), filepath.Ext(vcfPath)) + "_report.pdf"
	if err := generateHTMLReport(contactList, outputHTML); err != nil {
		log.Fatal("Ошибка генерации HTML:", err)
	}
	//if err := generatePDFReport(contactList, outputPDF); err != nil {
	//	log.Fatal("Ошибка генерации PDF:", err)
	//}

	fmt.Println("✅ HTML сохранён в:", outputHTML)
	//fmt.Println("✅ PDF сохранён в:", outputPDF)
}
