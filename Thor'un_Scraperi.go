package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Hedefler []string `yaml:"hedefler"`
}

func HedefOku(filepath string) ([]string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	//Dosyayı byte olarak oku
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	//YAML verisini struct'a işle
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return cfg.Hedefler, nil
}

func torHTTPistemci() (*http.Client, error) {
	dialer, err := proxy.SOCKS5(
		"tcp",
		"127.0.0.1:9150", // Tor varsayılan SOCKS5 portu
		nil,
		proxy.Direct,
	)

	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		Dial: dialer.Dial,
	}

	istemci := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	return istemci, nil
}

func EkrangoruntusuAl(url string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://127.0.0.1:9150"),
		chromedp.WindowSize(1920, 1080),
		chromedp.IgnoreCertErrors,
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	//taarayıcı context'i oluştur (Timeout ekleyelim ki takılmasın)
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	var buf []byte

	//chrome emir veriyoruz -> bekle -> foto çek
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return err
	}

	//Dosya kaydet
	os.MkdirAll("çıktı/ekrangoruntuleri", 0755)
	fileName := url
	fileName = strings.ReplaceAll(fileName, "https://", "")
	fileName = strings.ReplaceAll(fileName, "http://", "")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = fmt.Sprintf("%s.png", fileName)
	return os.WriteFile(filepath.Join("çıktı/ekrangoruntuleri", fileName), buf, 0644)
}

func fetch(wg *sync.WaitGroup, istemci *http.Client, url string) {
	defer wg.Done()

	if istemci == nil {
		fmt.Printf("HTTP istemcisi nil, istek atilmaz")
		return
	}

	resp, err := istemci.Get(url)
	if err != nil {
		durum := "HATA"
		if strings.Contains(err.Error(), "SOCKS5 proxy") {
			durum = "ONION_ERISIM_YOK"
			fmt.Printf("[UYARI] Onion servisine ulaşilamadi: %s\n", url)
		} else if strings.Contains(err.Error(), "timeout") {
			durum = "ZAMAN_ASIMI"
			fmt.Printf("[UYARI] Zaman asimi: %s\n", url)
		} else {
			fmt.Printf("[HATA] Tarama hatası: %s -> %s\n", url, err.Error())
		}
		appendLog(url, durum)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		appendLog(url, "OKUMA_HATASI")
		return
	}

	fmt.Printf("[BİLGİ] Veri Çekildi: %s -> BAŞARALI (%d bytes)\n", url, len(body))

	//HTML dosyasını kaydet
	err = HTMLkaydet(url, body)
	if err != nil {
		fmt.Printf("[HATA] HTML kaydediliyor: %s -> %s\n", url, err.Error())
		appendLog(url, "BAŞARISIZ")
	}
	//Ekran görüntüsü al
	fmt.Printf("[FOTO] Ekran görüntüsü alınıyor: %s\n", url)
	resimErr := EkrangoruntusuAl(url)
	if resimErr != nil {
		fmt.Printf("[UYARI] Ekran görüntüsü alınamadı %v\n", resimErr)
	} else {
		fmt.Printf("[FOTO] Ekran görünütüsü alındı: %s\n", url)
	}

	//Başarılı Log
	appendLog(url, "BAŞARILI")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Kullanim: go run . hedefler.yaml")
		return
	}
	hedefDosya := os.Args[1]
	hedefler, err := HedefOku(hedefDosya)
	if err != nil {
		fmt.Println("Hedef doyasi okunamadi:", err)
		return
	}

	fmt.Printf("[BASLAT] %d adet hedef yuklendi, tarama basliyor...\n", len(hedefler))

	istemci, err := torHTTPistemci()
	if err != nil {
		fmt.Println("Tor istemcisi oluşturulamadi:", err)
		return
	}

	var wg sync.WaitGroup
	fmt.Println("[SİSTEM] Goroutine (Eşzamanlı) tarama başlatılıyor...")

	for _, url := range hedefler {
		wg.Add(1) //her url için sayacı bir arttır
		// go komutu ile işlemi arka olana at
		go fetch(&wg, istemci, url)
	}
	wg.Wait() //hepsinin bitmesini bekle

	fmt.Println("[BİLGİ] Tor SOCKS proxy üzerinden istekler gönderiliyor")
}

func kontrolTorIP(istemci *http.Client) {
	url := "http://checktor.onion"

	resp, err := istemci.Get(url)
	if err != nil {
		fmt.Println("[HATA] tor IP kontrol iptal", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	content := string(body)

	if strings.Contains(content, "Tebrikler") || strings.Contains(content, "Congratulations") {
		fmt.Println("[Bilgi] Tor ağina başariyla bağlandiniz")
	} else {
		fmt.Println("[UYARI] Tor ağina bağlanilamadi!")
	}
}

func HTMLkaydet(url string, data []byte) error {
	// çıktı klasörünü oluştur
	os.Mkdir("çıktı/html", 0755)

	//URL'yi dosya adına çevir
	fileName := url
	fileName = strings.ReplaceAll(fileName, "https://", "")
	fileName = strings.ReplaceAll(fileName, "http://", "")
	fileName = strings.ReplaceAll(fileName, ":", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = fileName + ".html"

	filePath := filepath.Join("çıktı/html", fileName)

	return os.WriteFile(filePath, data, 0644)
}

func appendLog(url, status string) {
	os.Mkdir("çıktı", 0755)
	logFile := "çıktı/scan_report.log"

	entry := fmt.Sprintf("%s | %s | %s\n", time.Now().Format(time.RFC3339), url, status)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[ERR ]LOG Yaziliyor:", err)
		return
	}
	defer f.Close()
	f.WriteString(entry)
}
