# Thor's Scraper | Dark Web CTI Tool

**Thor's Scraper**, Siber Tehdit İstihbaratı (CTI) süreçlerinde kullanılmak üzere tasarlanmış, **Go (Golang)** tabanlı, yüksek performanslı bir .onion (Dark Web) tarama ve veri toplama aracıdır.

Bu proje; hedef listesindeki adresleri anonim olarak tarar, kaynak kodlarını indirir ve **Headless Browser** teknolojisi kullanarak sitelerin anlık ekran görüntülerini kaydeder.

---

## One Cikan Ozellikler

- **Yuksek Performans (Concurrency):** `Goroutines` ve `sync.WaitGroup` mimarisi sayesinde yüzlerce hedefi aynı anda, birbirini beklemeden tarar.
- **Tam Anonimlik:** Tüm HTTP ve Browser trafiği, özel yapılandırılmış `SOCKS5` istemcisi üzerinden **Tor Agina (127.0.0.1:9150)** yönlendirilir.
- **Gorsel Istihbarat:** Standart HTML indirmesinin ötesine geçerek, **`chromedp`** kütüphanesi ile siteleri render eder ve kanıt niteliğinde ekran görüntüsü (.png) alır.
- **Hata Toleransi:** Kapanmış veya erişilemeyen .onion siteleri programı durdurmaz; akıllı hata yönetimi ile raporlanır ve tarama devam eder.
- **Windows Uyumlu:** Dosya isimlendirme ve klasör yapısı, işletim sistemi kısıtlamalarına uygun olarak otomatik sanitize edilir.

---

## Kurulum ve Gereksinimler

Bu projeyi çalıştırmak için bilgisayarınızda aşağıdakilerin kurulu olması gerekir:

1.  **Go (Golang):** [Indir ve Kur](https://go.dev/dl/)
2.  **Tor Browser:** [Indir](https://www.torproject.org/download/) (Arka planda açık olmalı)
3.  **Google Chrome:** (Ekran görüntüsü alabilmek için gereklidir)

### Depoyu Klonlayin

```bash
git clone https://github.com/Polat174/Thor-Scraper.git
(https://github.com/Polat174/Thor-Scraper.git)
cd Thor-Scraper