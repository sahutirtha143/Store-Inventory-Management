
// // <!-- CODE FOR SECTIONS-->
function showSection(sectionId) {
    // Hide all sections
    document.querySelectorAll('.content').forEach((section) => {
        section.style.display = 'none';
    });

    // Show the selected section
    document.getElementById(sectionId).style.display = 'block';
}

// // Load the Buy section by default when the page loads
window.onload = function () {
    showSection('buy');
};

function handleFormSubmit(event, sectionId) {
    event.preventDefault(); // Prevent the default form submission

    // For demonstration, simply log the values to the console
    const formData = new FormData(event.target);
    console.log("Form submitted!", Object.fromEntries(formData));

    // Redirect to the appropriate section based on which form was submitted
    showSection(sectionId);

    // Optionally, you can refresh the inventory table or provide feedback to the user
    alert("Item added successfully!"); // Just an example message
}





// <!-- CODE FOR PROFIT LOSS VALUE -->

function calculateProfitLoss() {
    // Get the total buy and sell prices from the HTML template variables
    const totalBuyPrice = parseFloat("{{.TotalPriceAllItems}}");
    const totalSellPrice = parseFloat("{{.TotalPriceAllSelling}}");

    // Calculate the profit or loss
    const profitLoss = totalSellPrice - totalBuyPrice;

    // Update the displayed profit/loss value in the correct color
    const profitLossElement = document.getElementById('profitLoss');
    profitLossElement.innerHTML = `₹${profitLoss.toFixed(2)}`;

    // If it's profit (positive value), show it in green; if loss (negative), show it in red
    if (profitLoss >= 0) {
        profitLossElement.style.color = "green";
    } else {
        profitLossElement.style.color = "red";
    }
}

// Automatically calculate and display profit/loss after page load
window.onload = function () {
    calculateProfitLoss();
};


//DELETE:
function confirmDelete() {
    return confirm("Are you sure you want to delete this item?");
}



//<!-- CODE FOR PAGINATION -->

let currentPageBuy = 1;
let currentPageSell = 1;
const rowsPerPage = 5;

window.onload = function () {
    paginateTable('buy');
    paginateTable('sell');
};

function paginateTable(type) {
    let tableBody, currentPage;
    if (type === 'buy') {
        tableBody = document.getElementById('inventoryBodyBuy');
        currentPage = currentPageBuy;
    } else {
        tableBody = document.getElementById('inventoryBodySell');
        currentPage = currentPageSell;
    }

    const rows = tableBody.getElementsByTagName('tr');
    const totalRows = rows.length;
    const totalPages = Math.ceil(totalRows / rowsPerPage);

    // Hide all rows initially
    for (let i = 0; i < totalRows; i++) {
        rows[i].style.display = 'none';
    }

    // Display rows for the current page
    const startRow = (currentPage - 1) * rowsPerPage;
    const endRow = Math.min(startRow + rowsPerPage, totalRows);
    for (let i = startRow; i < endRow; i++) {
        rows[i].style.display = 'table-row';
    }

    // Update the page number display
    if (type === 'buy') {
        document.getElementById('pageNumberBuy').textContent = `Page ${currentPageBuy} of ${totalPages}`;
        document.getElementById('prevPageBuy').disabled = currentPageBuy === 1;
        document.getElementById('nextPageBuy').disabled = currentPageBuy === totalPages;
    } else {
        document.getElementById('pageNumberSell').textContent = `Page ${currentPageSell} of ${totalPages}`;
        document.getElementById('prevPageSell').disabled = currentPageSell === 1;
        document.getElementById('nextPageSell').disabled = currentPageSell === totalPages;
    }
}

function nextPage(type) {
    if (type === 'buy') {
        const totalRows = document.getElementById('inventoryBodyBuy').getElementsByTagName('tr').length;
        const totalPages = Math.ceil(totalRows / rowsPerPage);
        if (currentPageBuy < totalPages) {
            currentPageBuy++;
            paginateTable('buy');
        }
    } else {
        const totalRows = document.getElementById('inventoryBodySell').getElementsByTagName('tr').length;
        const totalPages = Math.ceil(totalRows / rowsPerPage);
        if (currentPageSell < totalPages) {
            currentPageSell++;
            paginateTable('sell');
        }
    }
}

function prevPage(type) {
    if (type === 'buy' && currentPageBuy > 1) {
        currentPageBuy--;
        paginateTable('buy');
    } else if (type === 'sell' && currentPageSell > 1) {
        currentPageSell--;
        paginateTable('sell');
    }
}











// // <!-- CODE FOR BOTH SEARCHING By TEXT and DATE -->

// Function to filter inventory based on search query or date
function filterInventory(section, filterType) {
    let filter;
    if (filterType === 'text') {
        const searchInput = document.getElementById('search' + section.charAt(0).toUpperCase() + section.slice(1));
        filter = searchInput.value.toLowerCase(); // Filter for text
    } else if (filterType === 'date') {
        const dateInput = document.getElementById('date' + section.charAt(0).toUpperCase() + section.slice(1));
        filter = dateInput.value; // Filter for date
    }

    // Get the correct table body for the section (Buy/Sell)
    const inventoryBody = document.getElementById('inventoryBody' + section.charAt(0).toUpperCase() + section.slice(1));
    const rows = inventoryBody.getElementsByTagName('tr');

    // Loop through all rows and hide those that don't match the search query or date
    for (let i = 0; i < rows.length; i++) {
        const cells = rows[i].getElementsByTagName('td');
        let match = false;

        // Check each cell in the row for a match
        for (let j = 0; j < cells.length; j++) {
            if (cells[j]) {
                const cellValue = cells[j].textContent || cells[j].innerText;

                // Handle text-based filtering
                if (filterType === 'text' && cellValue.toLowerCase().indexOf(filter) > -1) {
                    match = true;
                    break; // Stop checking if a match is found
                }

                // Handle date-based filtering (assuming the date is in the format yyyy-mm-dd)
                if (filterType === 'date' && filter !== "" && cellValue.includes(filter)) {
                    match = true;
                    break; // Stop checking if a match is found
                }
            }
        }

        // Show or hide the row based on the match
        rows[i].style.display = match ? "" : "none";
    }
}

// Event listener for search inputs (Buy and Sell)
document.getElementById('searchBuy').addEventListener('input', function () {
    filterInventory('Buy', 'text');
});

document.getElementById('searchSell').addEventListener('input', function () {
    filterInventory('Sell', 'text');
});

// Event listener for date inputs (Buy and Sell)
document.getElementById('dateBuy').addEventListener('input', function () {
    filterInventory('Buy', 'date');
});

document.getElementById('dateSell').addEventListener('input', function () {
    filterInventory('Sell', 'date');
});




// GRAPH
// Function to fetch inventory data from the server
async function fetchInventoryData() {
    const response = await fetch('/inventory-data');
    const data = await response.json();
    return data;
}

// Function to plot both charts using Chart.js
async function plotCharts() {
    const inventoryData = await fetchInventoryData();

    // Plot for the first chart (line chart)
    const labels = inventoryData.inventory.map(item => item.Date);
    const buyPrices = inventoryData.inventory.map(item => item.TotalPriceBuying);
    const sellPrices = inventoryData.selling_inventory.map(item => item.TotalPriceSelling);

    const ctx1 = document.getElementById('inventoryChart').getContext('2d');
    new Chart(ctx1, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [
                {
                    label: 'Total Buy Price',
                    data: buyPrices,
                    backgroundColor: 'rgba(0, 255, 0, 0.2)',
                    borderColor: 'green',
                    borderWidth: 2,
                    fill: false,
                    pointBackgroundColor: 'green',
                    tension: 0.4
                },
                {
                    label: 'Total Sell Price',
                    data: sellPrices,
                    backgroundColor: 'rgba(255, 0, 0, 0.2)',
                    borderColor: 'red',
                    borderWidth: 2,
                    fill: false,
                    pointBackgroundColor: 'red',
                    tension: 0.4
                }
            ]
        },
        options: {
            scales: {
                x: {
                    title: {
                        display: true,
                        text: 'Date'
                    }
                },
                y: {
                    beginAtZero: true,
                    title: {
                        display: true,
                        text: 'Price'
                    }
                }
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    labels: {
                        boxWidth: 20,
                        padding: 15
                    }
                }
            }
        }
    });

    // Plot for the second chart (bar chart)
    const itemLabels = inventoryData.inventory.map(item => item.Item);
    const buyPrices2 = inventoryData.inventory.map(item => item.TotalPriceBuying);
    const sellPrices2 = inventoryData.selling_inventory.map(item => item.TotalPriceSelling);

    const ctx2 = document.getElementById('inventoryChart2').getContext('2d');
    new Chart(ctx2, {
        type: 'bar',
        data: {
            labels: itemLabels,
            datasets: [
                {
                    label: 'Total Buy Price',
                    data: buyPrices2,
                    backgroundColor: 'rgba(75, 192, 192, 0.6)',
                    borderColor: 'rgba(75, 192, 192, 1)',
                    borderWidth: 1
                },
                {
                    label: 'Total Sell Price',
                    data: sellPrices2,
                    backgroundColor: 'rgba(153, 102, 255, 0.6)',
                    borderColor: 'rgba(153, 102, 255, 1)',
                    borderWidth: 1
                }
            ]
        },
        options: {
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

// Call the function to plot both charts when the page loads
window.onload = plotCharts;