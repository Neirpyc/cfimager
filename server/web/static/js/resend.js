function hcaptcha_ok(){
    document.getElementById("submit").disabled = false;
}

document.getElementById("form").addEventListener("submit", async function (event) {
    event.preventDefault();

    let obj = {
        "email": document.getElementById("email").value,
        "hcaptcha": hcaptcha.getResponse(),
    }

    fetch('/v1/resendEmail', {
        method: 'POST',
        body: JSON.stringify(obj)
    })
        .then(response => response.json())
        .then(data => {
            if (!data['success']) {
                errorLabel.innerHTML = data['error']
                hcaptcha.reset()
            } else {
                window.location.href = "../login"
            }
        })
        .catch(error => {
            console.log(error)
        })
});

const errorLabel = document.getElementById("error")