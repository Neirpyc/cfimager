async function hash(password) {
    return bytesArrToBase64((await argon2.hash({
        pass: password,
        salt: "vXA2vPiUeWjSzlx1kLFM9VyRXfIbSoBnTqVMQ2jlM5e6VW5hDCRPZTPWP49EsUcO",

        time: 1,
        mem: 1024 * 16,
        hashLen: 64,
        parallelism: 1,
        type: argon2.ArgonType.Argon2id,
    })).hash);
}

function hcaptcha_ok() {
    document.getElementById("submit").disabled = false;
}

function bytesArrToBase64(arr) {
    const abc = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
    const bin = n => n.toString(2).padStart(8, 0);
    const l = arr.length
    let result = '';

    for (let i = 0; i <= (l - 1) / 3; i++) {
        let c1 = i * 3 + 1 >= l;
        let c2 = i * 3 + 2 >= l;
        let chunk = bin(arr[3 * i]) + bin(c1 ? 0 : arr[3 * i + 1]) + bin(c2 ? 0 : arr[3 * i + 2]);
        let r = chunk.match(/.{1,6}/g).map((x, j) => j == 3 && c2 ? '=' : (j == 2 && c1 ? '=' : abc[+('0b' + x)]));
        result += r.join('');
    }

    return result;
}

document.getElementById("form").addEventListener("submit", async function (event) {
    event.preventDefault();
    errorLabel.innerHTML = ""

    let password = document.getElementById("password").value;
    let passwordRe = document.getElementById("passwordRe").value;
    if (password !== passwordRe) {
        document.getElementById("password").setCustomValidity("Passwords must match");
        return;
    }

    for (let i = 0; i < passwordRe.length; i++)
        passwordRe[i] = '@';

    let array = new Uint8ClampedArray(64);
    window.crypto.getRandomValues(array);

    let obj = {
        "email": document.getElementById("email").value,
        "password": await hash(document.getElementById("password").value),
        "hcaptcha": hcaptcha.getResponse(),
        "salt": bytesArrToBase64(array)
    }

    for (let i = 0; i < password.length; i++)
        password[i] = '@';

    fetch('/v1/register', {
        method: 'POST',
        body: JSON.stringify(obj)
    })
        .then(response => response.json())
        .then(data => {
            if (!data['success']) {
                errorLabel.innerHTML = data['error']
                if (data['error'] == "EMAIL_NOT_VALIDATED") {
                    window.location = "../postregister?email=" + obj['email'];
                } else if (data['error'] == "LOGIN") {
                    window.location = "../login";
                }
                hcaptcha.reset()
            } else {
                window.location = "../postregister/?email=" + obj['email']
            }
        })
        .catch(error => {
            console.log(error)
        })
});

const errorLabel = document.getElementById("error")
