document.querySelectorAll("#stats tbody").forEach(function(tbody) {
    var max = new Array();

    tbody.querySelectorAll("tr").forEach(function(tr) {
        tr.querySelectorAll("td.num").forEach(function(td, i) {
            if (!max[i]) {
                max[i] = 0;
            }
            var v = parseFloat(td.innerText);
            if (v > max[i]) {
                max[i] = v;
            }
        });
    });

    tbody.querySelectorAll("tr").forEach(function(tr) {
        tr.querySelectorAll("td.num").forEach(function(td, i) {
            if(max[i] == parseFloat(td.innerText)) {
                td.classList.add("max");
            }
        });
    });
});

document.querySelectorAll("th").forEach(function(th) {
    if(th.colSpan == 1) {
        th.classList.add("sortable");
        th.onclick = sortTH;
    }
});

function sortTH(evt) {
    const self = this;
    const tbl = this.parentNode.parentNode.parentNode;
    var col = 0;

    if(self.classList.contains("sort-asc")) {
        self.classList.remove("sort-asc");
        self.classList.add("sort-desc");
    } else {
        self.classList.remove("sort-desc");
        self.classList.add("sort-asc");
    }

    const asc = self.classList.contains("sort-asc");

    tbl.querySelectorAll("th.sortable").forEach(function(th, i) {
        th.classList.remove("sort-asc");
        th.classList.remove("sort-desc");
        if(th == self) {
            th.classList.add(asc ? "sort-asc" : "sort-desc");
            col = i
        };
    });

    const trs = tbl.querySelectorAll("tbody tr");
    var sortedTrs = Array.prototype.slice.call(trs, 0);

    sortedTrs.sort(function(a, b) {
        var va = parseFloat(a.getElementsByTagName("td").item(col).innerText);
        var vb = parseFloat(b.getElementsByTagName("td").item(col).innerText);

        if(isNaN(va)) va = a.getElementsByTagName("td").item(col).innerText;
        if(isNaN(vb)) vb = b.getElementsByTagName("td").item(col).innerText;

        if(asc) {
            if(va > vb) return 1;
            if(va < vb) return -1;
        } else {
            if(va > vb) return -1;
            if(va < vb) return 1;
        }
        return 0;
    });

    const tbody = tbl.getElementsByTagName("tbody").item(0);

    tbody.querySelectorAll("tr").forEach(function(tr) {
        tbody.removeChild(tr);
    });

    sortedTrs.forEach(function(tr) {
        tbody.appendChild(tr);
    });
}

const body = document.body;
const ol = document.createElement("ol");
ol.id = "toc";
body.insertBefore(ol, body.firstElementChild.nextElementSibling);

document.querySelectorAll("h2").forEach(function(h) {
    const id = h.parentNode.id;
    const li = document.createElement("li");
    const a = document.createElement("a");
    a.href = "#" + id;
    a.innerText = h.innerText;
    li.appendChild(a);
    ol.appendChild(li);
});
