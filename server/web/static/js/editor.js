function hcaptcha_ok() {
    document.getElementById("submit").disabled = false;
}

function updateURLParameter(url, param, paramVal) {
    var TheAnchor = null;
    var newAdditionalURL = "";
    var tempArray = url.split("?");
    var baseURL = tempArray[0];
    var additionalURL = tempArray[1];
    var temp = "";

    if (additionalURL) {
        var tmpAnchor = additionalURL.split("#");
        var TheParams = tmpAnchor[0];
        TheAnchor = tmpAnchor[1];
        if (TheAnchor)
            additionalURL = TheParams;

        tempArray = additionalURL.split("&");

        for (var i = 0; i < tempArray.length; i++) {
            if (tempArray[i].split('=')[0] != param) {
                newAdditionalURL += temp + tempArray[i];
                temp = "&";
            }
        }
    } else {
        var tmpAnchor = baseURL.split("#");
        var TheParams = tmpAnchor[0];
        TheAnchor = tmpAnchor[1];

        if (TheParams)
            baseURL = TheParams;
    }

    if (TheAnchor)
        paramVal += "#" + TheAnchor;

    var rows_txt = temp + "" + param + "=" + paramVal;
    return baseURL + "?" + newAdditionalURL + rows_txt;
}

document.getElementById("form").addEventListener("submit", async function (event) {
    event.preventDefault();
    errorLabel.innerHTML = ""

    let obj = {
        "hcaptcha": hcaptcha.getResponse(),
        "id": parseInt(new URLSearchParams(window.location.search).get("id"), 10)
    };

    const newName = name.value;
    let nameModified = false;
    if (newName != funcName) {
        obj['newname'] = newName;
        nameModified = true;
    }
    const newContent = content.value;

    if (newContent != oldContent) {
        obj['modified'] = true;
        obj['content'] = newContent;
    }

    errorLabel.innerHTML = "Compiling... This can take up to a minute";
    fetch('../v1/editFunction', {
        method: 'POST',
        body: JSON.stringify(obj)
    })
        .then(response => response.json())
        .then(data => {
            if (!data['success']) {
                errorLabel.innerHTML = data['error'];
                errorArea.innerHTML = data['message'];
                content.value = data['content'] != undefined ? data['content'] : content.value;
                funcName = data['name'] != undefined ? data['name'] : name.value;
            } else {
                errorArea.innerHTML = "Success"
                errorLabel.innerHTML = "Success";
            }
            if (name.value != funcName) {
                name.value = funcName;
                window.history.replaceState("", "", updateURLParameter(window.location.href, "name", name.value));
            }
            hcaptcha.reset()
        })
        .catch(error => {
            console.log(error)
        })

});

const errorArea = document.getElementById("errorArea");
const errorLabel = document.getElementById("error");
let funcName = new URLSearchParams(window.location.search).get("name");
const oldContent = document.getElementById("code").value;
const content = document.getElementById("code");
const name = document.getElementById("name");