/*
 * @brief   main.js
 * @author  Sebastien LEGRAND
 * @date    2023-03-08
 *
 * @brief   Main Javascript file for "Bouffe Action" collector
 */

//----- constants

// flask endpoint
const SERVER_URL="http://127.0.0.1:5000/api/v1"

const SCALE_TIMER_INTERVAL_MS = 2000;
const INPUT_FOCUS_INTERVAL_MS = 1000;

const BARCODE_PROVIDER_MARKER = 'F';
const BARCODE_PRODUCT_MARKER = 'P';

const QUANTITY_MULTIPLIER_UPPER = 'X'
const QUANTITY_MULTIPLIER_LOWER = 'x'
const QUANTITY_MULTIPLIER_SYMB = '*'

const COMMENT_MARKER = '#'

const DEFAULT_QUANTITY = 1


//----- globals

var last_provider = "";
var last_weight = 0.0;
var last_items = 1;
var last_element_id = 0;


//----- functions
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function animateEntry(ms) {
    var entry = document.getElementById("barcode-entry");
    var old_style = entry.className;

    entry.className = entry.className + " bg-danger text-white";
    sleep(ms).then(() => {
        entry.className = old_style;
    });
}

// setup the page
function setup() {
    // setup the scale monitoring every seconds or so
    setInterval(readScaleValue, SCALE_TIMER_INTERVAL_MS);
    setInterval(setInputFocus, INPUT_FOCUS_INTERVAL_MS);
    fetchExistingEntries();

    // bind the enter key to the barcode button
    window.addEventListener("keypress", (event) => {
        if (event.key == "Enter") {
            document.getElementById("barcode-btn").click();
        }
    });
}

// reset focus to input field
function setInputFocus() {
    var inputBar = document.getElementById("barcode-entry");
    if(document.activeElement != inputBar) {
        inputBar.focus()
    }
}


// read the scale value
function readScaleValue() {
    fetch(SERVER_URL + "/scale")
    .then((response) => {
        return response.json();
    })
    .then((json) => {
       var element = document.getElementById("current-weight");
       element.innerHTML = json['value'];
       last_weight = json['value'];
    })
    .catch(() => {
        console.log("Unable to retrieve the value of the scale!");
    });
}

// read the entry from the user
function readEntry() {
    var entry = document.getElementById("barcode-entry");

    switch (entry.value.charAt(0)) {
        case BARCODE_PROVIDER_MARKER:
            retrieveProvider(entry.value);
            break;
        case BARCODE_PRODUCT_MARKER:
            retrieveProduct(entry.value);
            break;
        case QUANTITY_MULTIPLIER_UPPER:
        case QUANTITY_MULTIPLIER_LOWER:
        case QUANTITY_MULTIPLIER_SYMB:
            processMultiplier(entry.value)
            break;
        case COMMENT_MARKER:
            processComment(entry.value)
            break;
        default:
            console.log('Meh?');
            break;
    }
}

// retrieve the entity name
function retrieveProvider(eid) {
    fetch(SERVER_URL + "/entity/" + eid)
    .then((response) => {
        if (response.status != 200) {
            // do a little animation to notify the user
            animateEntry(200);
            throw 'cancel';
        } else {
            return response.json();
        }
    })
    .then((json) => {
        var provider = document.getElementById("last-provider");
        provider.innerHTML = json['ename'];
        last_provider = json['ename'];

        var entry = document.getElementById("barcode-entry");
        entry.value = "";
    })
    .catch(() => {
        return "";
    });
}

// retrieve the product
function retrieveProduct(eid) {
    fetch(SERVER_URL + "/entity/" + eid)
    .then((response) => {
        if (response.status != 200) {
            // do a little animation to notify the user
            animateEntry(200);
            throw 'cancel';
        } else {
            return response.json();
        }
    })
    .then((json) => {
        var entry = document.getElementById("barcode-entry");

        if (last_provider == "") {
            entry.value = "";
            return;
        }

        addTableEntry(last_provider, json['ename'], last_weight);
        entry.value = "";
    })
    .catch(() => {
        return "";
    });

}

function addHTMLEntry(id, provider, product, quantity, weight) {
    var table = document.getElementById("table-content");
    var x = table.rows.length;
    var row = table.insertRow(1);

    // add the data to the table
    row.id = id;
    row.insertCell(0).innerHTML = `<th scope="row">${id}</th>`;
    row.insertCell(1).innerHTML = provider;
    row.insertCell(2).innerHTML = product;
    row.insertCell(3).innerHTML = quantity;
    row.insertCell(4).innerHTML = weight;
    row.insertCell(5).innerHTML = '<input type="button" class="btn btn-danger btn-sm" id="delrow" onclick="deleteRow(\'' + id + '\')" value="Supprimer" />';
    row.insertCell(6).innerHTML = '';

    // increment the number of items
    last_items = last_items + 1;

    // record the last element ID (to apply multiplier if there is any)
    last_element_id = id;
}

// add a new entry in the table
function addTableEntry(provider, product, weight) {
    // send the item to the backend
    fetch(SERVER_URL + "/input", {
        method: "POST",
        body: JSON.stringify({
            provider: provider,
            product: product,
            quantity: DEFAULT_QUANTITY,
            weight: weight
        }),
        headers: {
            "Content-Type": "application/json; charset=UTF-8"
        }
    })
    .then((response) => {
        return response.json()
    })
    .then((json) => {
        // the json contains the ID of the item and the weight
        // weight can be changed if the product = compost
        var id = json['id']
        weight = json['weight']

        // create a new entry in the table
        addHTMLEntry(id, provider, product ,DEFAULT_QUANTITY, weight);
    });
}

function fetchExistingEntries() {
    fetch(SERVER_URL + "/input", {
        method: "GET",
        headers: {
            "Content-Type": "application/json; charset=UTF-8"
        }
    })
    .then((response) => {
        return response.json()
    })
    .then((json) => {
        for(var key in json) {
            entry = json[key]
            console.log(entry)
            addHTMLEntry(entry.id, entry.provider, entry.product ,entry.quantity, entry.weight);
        }
    })
}

function processMultiplier(value) {
    // clean the input and create a number out of it
    value = value.replace('x', '').replace('X', '').replace('*', '').replace(' ', '');

    if (!isNaN(value)) {
        fetch(SERVER_URL + '/input/' + last_element_id, {
            method: "PUT",
            body: JSON.stringify({
                quantity: Number(value),
            }),
            headers: {
                "Content-Type": "application/json; charset=UTF-8"
            }
        })
        .then((response) => {
            document.getElementById(last_element_id).cells[3].innerHTML = value

            // empty the barcode entry line on success
            var entry = document.getElementById("barcode-entry");
            entry.value = "";
        })
    } else {
        animateEntry(200);
    }
}

function processComment(value) {
    // clean the input
    value = value.replace('#', '');

    // send the comment to the backend
    fetch(SERVER_URL + '/input/' + last_element_id, {
        method: "PUT",
        body: JSON.stringify({
            comment: value
        }),
        headers: {
            "Content-Type": "application/json; charset=UTF-8"
        }
    })
    .then((response) => {
        // add a clip with the message as a tooltip
        text  = '<i class="fa-solid fa-paperclip" '
        text += 'data-bs-toggle="tooltip" data-bs-placement="top" '
        text += 'title="' + value + '">&nbsp</i>'
        document.getElementById(last_element_id).cells[6].innerHTML = text

        // empty the barcode entry line
        var entry = document.getElementById("barcode-entry");
        entry.value = "";
    })
}

function deleteRow(id) {
    /* remove the item from the backend first */
    fetch(SERVER_URL + "/input/" + id, {
        method: "DELETE"
    })
    .then((response) => {
        return response.json()
    })
    .then((json) => {
        document.getElementById(id).remove();
    });
}

function downloadXLS() {
    var filename = "";

    fetch(SERVER_URL + "/download")
    .then((response) => {

        // determine the filename from the content dispositon header
        var cd = response.headers.get("content-disposition");
        var regex = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/;
        var match = regex.exec(cd);
        filename = match[1] || "data_export.xlsx";

        // return the remaining of the response as a "blob"
        return response.blob();
    })
    .then((blob) => {
        try {
            //downloading the file depends on the browser
            //IE handles it differently than chrome/webkit
            if (window.navigator && window.navigator.msSaveOrOpenBlob) {
                window.navigator.msSaveOrOpenBlob(blob, filename);
            } else {
                // create an object URL from the blob
                var url = URL.createObjectURL(blob);

                // create a "a href" link, and click it
                var a = document.createElement("a");
                a.href = url;
                a.download = filename;
                document.body.appendChild(a);
                a.click();

                // remove the "a href" as soon as possible
                setTimeout(() => {
                    document.body.removeChild(a);
                    window.URL.revokeObjectURL(url);
                }, 0);
            }
        } catch (exc) {
            console.log("Save Blob method failed with the following exception.");
            console.log(exc);
        }
    })
}

//----- begin
// activate tooltips
var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'))
var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
  return new bootstrap.Tooltip(tooltipTriggerEl)
})

// run setup
setup();