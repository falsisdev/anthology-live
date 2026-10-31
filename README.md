# 🚀 Anthology Live (v2 Go)

Anthology Canlı Yayın & Spor Motoru tabanlı; **Smart TV (Android TV, Apple TV, webOS, Tizen), VLC, IINA, Kodi, Telefon, Tablet ve Tarayıcılar** için sıfır kasma, sıfır donma ve tek parça livestream sağlayan bağımsız IPTV ve Xtream Codes proxy servisi.

Go (Golang) mimarisiyle yeniden yazılmış olup **`anthology-dav`** ile aynı SQLite veritabanını paylaşır.

---

## 📺 Kullanım Rehberi (`anthology-live.duckdns.org`)

### 🔑 Ortak Hesap Bilgileri
> [!IMPORTANT]
> **Anthology Live**, üst klasördeki **`anthology-dav`** projesi ile aynı kullanıcı veritabanını (`anthology.db`) kullanır.
> `anthology-dav` üzerinde kayıtlı olan mevcut kullanıcı adı ve şifreniz burada da **birebir geçerlidir**. Yeni bir hesap açmanıza gerek yoktur.

---

### 1. 📺 Smart TV'lerde İzleme (Android TV, Apple TV, webOS, Tizen)
Televizyonlarda web tarayıcı kullanmak zahmetli olduğu için sistem **Xtream Codes API** protokolünü destekler.

1. Smart TV'nize veya TV Box'ınıza popüler bir IPTV uygulaması kurun:
   - **Android TV / Google TV / FireStick:** [TiviMate](https://tivimate.com/), [Televizo](https://televizo.net/), [IPTV Smarters Pro](https://www.iptvsmarters.com/), XCIPTV.
   - **Apple TV / iOS:** [SenPlayer](https://apps.apple.com/), [Infuse](https://firecore.com/infuse), [Smarters Player Lite](https://apps.apple.com/).
   - **Samsung (Tizen) / LG (webOS):** IBO Player, Nanomid, SS IPTV.
2. Uygulamayı açıp **"Xtream Codes API ile Giriş Yap"** (Add Playlist > Xtream Codes) seçeneğini seçin.
3. Aşağıdaki bilgileri doldurun:
   - **Çalma Listesi Adı:** `Anthology Live`
   - **Sunucu Adresi (Server URL):** `https://anthology-live.duckdns.org`
   - **Kullanıcı Adı (Username):** `<anthology-dav kullanıcı adınız>`
   - **Şifre (Password):** `<anthology-dav şifreniz>`
4. **Giriş Yap** butonuna bastığınızda tüm spor ve ulusal kanallar kategorileri, EPG program rehberi ve logolarıyla otomatik olarak yüklenecektir.

---

### 2. 💻 VLC, IINA ve Harici Oynatıcılarda İzleme (PC / Mac)
Herhangi bir IPTV uygulaması kurmadan doğrudan VLC veya IINA ile izleyebilirsiniz:

1. **VLC Media Player'ı açın.**
2. **Medya (Media) > Ağ Akışı Aç (Open Network Stream)** yolunu izleyin (Kısayol: `Ctrl+N` veya Mac'te `Cmd+N`).
3. Aşağıdaki adresi yapıştırıp **Oynat** butonuna basın:
   ```text
   https://anthology-live.duckdns.org/iptv.m3u?username=KULLANICI_ADINIZ&password=SIFRENIZ
   ```
4. Kanallar listesini açarak (`Ctrl+L` / `Cmd+L`) dilediğiniz beIN Sports, S Sport, Tivibu veya TRT kanalını seçin.

---

### 3. 🌐 Web TV Portalı (Telefon, Tablet ve Tarayıcılar)
Hiçbir uygulama yüklemeden doğrudan web tarayıcınızdan izlemek için:

1. Tarayıcınızda şu adresi açın:
   ```text
   https://anthology-live.duckdns.org/
   ```
2. Şık Web TV arayüzünden izlemek istediğiniz kanala tıklayın.
3. Dahili **Hls.js** oynatıcı sayesinde yayın otomatik olarak tam ekran ve 1080p kalitesinde başlar.
4. İsterseniz kanalın altındaki **"⚡️ VLC ile Aç"** butonuna basarak yayını tek tıkla bilgisayarınızdaki VLC oynatıcısına aktarabilirsiniz.

---

### 4. 🔀 Çoklu Kaynak Seçimi (Canlı Yayın Kaynağını Değiştirme)
Önemli derbilerde veya yoğun maç günlerinde bir yayın platformu yoğunluktan takılırsa, yayını anında başka bir sağlayıcıya geçirebilirsiniz:

- **Web TV Üzerinden:**
  Oynatıcının hemen altında yer alan **"📡 Yayın Kaynağı"** bölümünden dilediğiniz kaynağa (`⚡️ Otomatik`, `Alemdar Sport`, `Selçuk / Arda`, `Taraftarium24`, `Mahsun Sports`, `Papaz Sports`) tıklayarak yayını anında kesintisiz başka bir kaynaktan devam ettirebilirsiniz.
- **Smart TV ve VLC için Alternatif Kanallar Listesi:**
  Eğer TV kumandanızdan direkt kanal değiştirir gibi kaynak değiştirmek isterseniz, M3U linkinizin sonuna `&all_sources=1` ekleyin:
  ```text
  https://anthology-live.duckdns.org/iptv.m3u?username=KULLANICI_ADI&password=SIFRE&all_sources=1
  ```
  Bu sayede çalma listenizde hem ana kanallar (otomatik en hızlı çalışan), hem de her kanalın **[Alemdar]**, **[Selçuk]**, **[Taraftarium]** gibi alternatifleri ayrı kanallar olarak listelenir.
- **Doğrudan Akış URL'si ile:**
  Herhangi bir kanal linkine `?source=alemdarsport` veya `?source=ardasports` ekleyerek istediğiniz kaynaktan açabilirsiniz:
  ```text
  https://anthology-live.duckdns.org/live/beinsports1.m3u8?source=alemdarsport
  ```

---

## ⚡️ Öne Çıkan Özellikler

1. **5 Büyük Canlı Platform Çözücü:**
   - AlemdarSport, ArdaSports (SelcukSports), Taraftarium24, MahsunSports ve PapazSports üzerinden canlı maç linklerini dinamik olarak çözer.
2. **Kendi Kendini Onaran Yayınlar (Self-Healing & Auto-Recovery):**
   - Arka planda çalışan goroutine nöbetçi her 15 saniyede bir kanalların sağlık durumunu test eder.
   - Domain değiştiğinde veya BTK engeli geldiğinde sistem yeni adresi otomatik yakalar; kesinti yaşatmadan çalışan yedek platforma geçer.
3. **Akıllı RAM Önbelleği (Go Restreamer Cache):**
   - 50 kişi aynı maçı izlese bile sunucu akışı kaynaktan **yalnızca 1 kez** çeker ve RAM tamponundan tüm izleyicilere sıfır gecikmeyle dağıtır. Bant genişliği ve sunucu CPU tüketimi asgaride kalır.
4. **Sıfır Donma & Otomatik MIME Type Düzeltme:**
   - CDN'ler üzerinden `.png` veya `.jpg` maskesiyle gelen TS video segmentlerini `video/mp2t` formatına dönüştürerek oynatıcının takılmasını veya donmasını %100 engeller.
5. **CORS ve Referer Tünelleme:**
   - CDN'lerin talep ettiği `Referer`, `Origin` ve `User-Agent` başlıklarını yerel olarak ekleyerek 403 erişim engellerini aşar.

---

## 🛠 Geliştirici & Yerel Kurulum Rehberi

Projeyi kendi bilgisayarınızda veya kendi VPS sunucunuzda derleyip çalıştırmak için:

### 1. Gereksinimler
- [Go (Golang)](https://go.dev/) 1.22 veya daha yenisi

### 2. Projeyi İndirme ve Doğrudan Çalıştırma
```bash
# Projeyi derlemeden doğrudan çalıştırmak için:
make run
```

### 3. macOS İçin Derleme ve Başlangıç Servisi Olarak Kurma
```bash
# macOS için ikili dosya (binary) derlemek için:
make build

# macOS başlangıç servisi (LaunchAgent) olarak kurmak ve hemen arka planda başlatmak için:
make install-service

# Servis durumunu kontrol etmek için:
make status

# Canlı logları izlemek için:
make logs

# Başlangıç servisinden kaldırmak ve durdurmak için:
make uninstall-service
```

### 4. Linux (AMD64) VPS İçin Derleme (Sunucu Dağıtımı)
Sunucunuzda Go kurulu olmasına gerek kalmadan tek dosyalık statik binary derleyebilirsiniz:
```bash
# CGO gerektirmeyen bağımsız Linux binary derler:
make build-linux
```
Oluşan **`anthology-live-linux`** (yaklaşık 11 MB) dosyasını VPS'inize `anthology-dav` ile aynı klasöre atıp çalıştırabilirsiniz:
```bash
chmod +x anthology-live-linux
./anthology-live-linux
```

### 5. Ortam Değişkenleri (Opsiyonel)
| Değişken | Varsayılan Değer | Açıklama |
| :--- | :--- | :--- |
| `PORT` | `3555` | Sunucunun dinleyeceği HTTP portu |
| `DB_PATH` | `../anthology-dav/data/anthology.db` | Ortak SQLite veritabanının dosya yolu |
| `BASE_URL` | `http://<yerel-ip>:3555` | Çalma listelerinde kullanılacak harici erişim URL'si |
| `SECRET_KEY` | *(Boş)* | Tanımlanırsa yönetici bypass anahtarı (`?token=...`) olarak kullanılır |

---

## 📄 Lisans
Bu proje MIT lisansı altındadır.
