# Proxy Server for Bank Payment Gateway

This is a **proxy server** designed to securely forward requests from a specified source to a payment gateway. It
validates requests using a public/private key mechanism and then forwards the data to a target bank server (such as
Iranian banks via Shaparak). This is necessary due to restrictions on foreign servers, and it acts as an intermediary to
ensure requests are properly accepted by local banks.

The proxy supports all banks that are members of **Shaparak** (the Iranian national payment gateway network).

---

## Features

- **Signature Verification**: Ensures that the requests come from a trusted source by verifying the signature of the
  data.
- **Data Decryption**: The `sec` field is encrypted and decrypted using RSA private/public keys to validate the
  integrity of the request.
- **Forward Requests**: After verification and decryption, the request is forwarded to the target bank's server for
  processing.
- **Supports Shaparak Banks**: The proxy is compatible with all Iranian banks that are members of the Shaparak network.

---

## Requirements

- Go 1.18+ for compiling the application.
- Public and private RSA keys for secure communication.
- A valid Shaparak-based bank's API endpoint.

---

## Setup

1. **Clone the Repository**

    ```bash
    git clone https://github.com/freecyberhawk/sep_proxy.git
    cd sep_proxy
    ```

2. **Build the Application**

   Compile the Go application:

    ```bash
    go build -o sep_proxy main.go
    ```

3. **Configure Keys**

   Place your RSA public and private key files in the root directory:
    - `public_key.pem` (for signature verification)

4. **Run the Proxy Server**

   Start the server with:

    ```bash
    ./sep_proxy
    ```

   By default, the server will run on `http://localhost:8080`.

---

## Usage

The proxy expects requests to be made to the following endpoint format:

```
http://<sep_proxy>/path?data=<data>&sec=<signature>&secval=<decrypted-sec-value>
```

- `sec`: The signature of the data.
- `secval`: The decrypted `sec` value.

The server will verify the signature, decrypt the `sec` field using the private key, and forward the request to the
target bank server if the signature is valid.

---

## Error Handling

- **Signature Verification Failure**: If the signature is invalid, the request is rejected with a `401 Unauthorized`
  error.
- **Decryption Failure**: If the `sec` field cannot be decrypted, the request is rejected with a `401 Unauthorized`
  error.
- **Forwarding Failure**: If the request cannot be forwarded to the bank server, a `502 Bad Gateway` error is returned.

---

## Persian (فارسی)

این یک **سرور پروکسی** است که برای هدایت درخواست‌ها از یک منبع مشخص به درگاه پرداخت بانکی طراحی شده است. این سرور
درخواست‌ها را با استفاده از مکانیسم کلید عمومی/خصوصی بررسی کرده و پس از اعتبارسنجی، داده‌ها را به سرور درگاه بانکی مقصد
ارسال می‌کند. این نیاز به دلیل محدودیت‌های موجود برای سرورهای خارجی است و به عنوان یک واسطه عمل می‌کند تا درخواست‌ها به
درستی توسط بانک‌های داخلی پذیرفته شوند.

این پروکسی با تمامی بانک‌های عضو **شاپرک** (شبکه درگاه پرداخت ملی ایران) سازگار است.

---

## ویژگی‌ها

- **اعتبارسنجی امضا**: اطمینان از اینکه درخواست‌ها از منبع مورد اعتماد می‌آیند با بررسی امضای داده‌ها.
- **رمزگشایی داده‌ها**: فیلد `sec` رمزگذاری شده و با استفاده از کلید خصوصی/عمومی RSA برای اعتبارسنجی داده‌ها رمزگشایی
  می‌شود.
- **هدایت درخواست‌ها**: پس از اعتبارسنجی و رمزگشایی، درخواست به سرور بانک مقصد هدایت می‌شود.
- **پشتیبانی از بانک‌های شاپرک**: پروکسی با تمامی بانک‌های ایرانی که عضو شبکه شاپرک هستند، سازگار است.

---

## الزامات

- Go 1.18+ برای کامپایل برنامه.
- کلیدهای RSA عمومی و خصوصی برای ارتباطات امن.
- یک نقطه پایان معتبر API بانک‌های عضو شاپرک.

---

## راه‌اندازی

1. **کلون کردن مخزن**

    ```bash
    git clone https://github.com/freecyberhawk/sep_proxy.git
    cd sep_proxy
    ```

2. **ساخت برنامه**

   برنامه Go را کامپایل کنید:

    ```bash
    go build -o sep_proxy main.go
    ```

3. **پیکربندی کلیدها**

   فایل‌های کلید عمومی و خصوصی RSA را در دایرکتوری ریشه قرار دهید:
    - `public_key.pem` (برای اعتبارسنجی امضا)

4. **اجرای سرور پروکسی**

   سرور را با دستور زیر راه‌اندازی کنید:

    ```bash
    ./sep_proxy
    ```

   به‌طور پیش‌فرض، سرور روی `http://localhost:8080` اجرا خواهد شد.

---

## استفاده

پروکسی انتظار دارد که درخواست‌ها به فرمت زیر ارسال شوند:

```
http://<sep_proxy>/path?data=<data>&sec=<signature>&secval=<decrypted-sec-value>
```

- `sec`: امضای داده‌ها.
- `secval`: مقدار رمزگشایی شده `sec`.

سرور امضا را بررسی کرده و فیلد `sec` را با استفاده از کلید خصوصی رمزگشایی می‌کند و در صورت معتبر بودن امضا، درخواست را
به سرور بانک مقصد ارسال می‌کند.

---

## مدیریت خطا

- **شکست در اعتبارسنجی امضا**: اگر امضا نامعتبر باشد، درخواست با خطای `401 Unauthorized` رد می‌شود.
- **شکست در رمزگشایی**: اگر فیلد `sec` نتواند رمزگشایی شود، درخواست با خطای `401 Unauthorized` رد می‌شود.
- **شکست در ارسال درخواست**: اگر درخواست نتواند به سرور بانک ارسال شود، خطای `502 Bad Gateway` برگردانده می‌شود.
