
const SERVER_URL="http://127.0.0.1:5000/api/v1"


function printType(entityType) {
    fetch(SERVER_URL + "/entitiesByType/"+entityType)
    .then((response) => {
        return response.json();
    })
    .then((json) => {
        var prevCategory = "";
        var root = getRoot();
        var currentCol = 0;
        var maxCol = 2;
        var currentTable = null;
        var currentLine = null;
        for(var key in json) {
            entry = json[key];
            if ((prevCategory != "") && (entry.category != prevCategory))
            {
                // except for the first one jump page whenever we change category
                var iDiv = addPageBreak(root);
            }
            // // we're starting a new category
            if ((prevCategory == "") || (entry.category != prevCategory)) {
                var irDiv = document.createElement('div')
                irDiv.className = "row border-bottom"
                var iDiv = document.createElement('div');
                iDiv.className = "barcode-header";
                iDiv.innerText = entry.category;
                irDiv.appendChild(iDiv)
                root.appendChild(irDiv);
                
                prevCategory = entry.category;

                currentTable =  document.createElement('table');
                currentTable.className = "table table-striped align-middle"
                root.appendChild(currentTable);
                console.log(" ==> Starting new table")
                currentCol = 0;

            }

            if ((currentCol == 0) || (currentCol > maxCol)) {
                console.log(" ==> Adding new line")
                currentLine =  currentTable.insertRow();
                currentCol = 0;
            }

            var cell = document.createElement('td');
            cell.className = "barcode-line";

            var code = document.createElement('div');
            code.className = "barcode";
            code.innerText = "*" + entry.code + "*";
            
            var name = document.createElement('div');
            name.className = "barcode-text";
            name.innerText = entry.name;
            
            cell.appendChild(code);
            cell.appendChild(name);
            console.log(cell.innerHTML)
            currentLine.appendChild(cell);

            currentCol +=1;
            
            console.log(entry)
         }
         var iDiv = addPageBreak(root);
    //    console.log(root.innerHTML)
    })
    .catch(() => {
        console.log("Unable to retrieve the value of the scale!");
    });
}

function addPageBreak(root) {
    var iDiv = document.createElement('div');
    iDiv.className = "pagebreak";
    root.appendChild(iDiv);
    // console.log("Adding line break before " + entry.category);
    return iDiv;
}

function getRoot() {
    return document.getElementById("print-section");
}

// setup the page
function print() {
    printType("F");
    printType("P");
}

print();