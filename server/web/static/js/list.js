window.onload = function(){
    const errorLabel = document.getElementById("error")
    document.getElementById("form").addEventListener("submit", async function (event) {
        event.preventDefault();
        errorLabel.innerHTML = ""

        let obj = {
            "name": document.getElementById("name").value,
        }


        fetch('/v1/createFunction', {
            method: 'POST',
            body: JSON.stringify(obj)
        })
            .then(response => response.json())
            .then(data => {
                if (!data['success']) {
                    errorLabel.innerHTML = data['error']
                    hcaptcha.reset()
                    return;
                }
                document.location.href = "../editor?name=" + obj['name'] + "&id=" + data['error'];
            })
            .catch(error => {
                console.log(error)
            })

    });
}

function deleteId(id, name) {
    if (confirm("Are you sure you want to delete the function " + name + "? This cannot be undone!") === true) {
        fetch('/v1/delete', {
            method: 'POST',
            body: id
        })
            .then(response => response.json())
            .then(data => {
                if (!data['success']) {
                    alert("Could not delete function " + name + " :" + data['error'])
                    return;
                }
                alert("Sucessfully deleted " + name + ".")
                location.reload(false)
            })
            .catch(error => {
                console.log(error)
            })
    }
}