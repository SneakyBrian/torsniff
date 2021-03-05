// import { sayHello } from "./greet";
// function showHello(divName: string, name: string) {
//   const elt = document.getElementById(divName);
//   elt.innerText = sayHello(name);
// }
// showHello("greeting", "TypeScript");

const button = document.getElementById("searchButton");
button.addEventListener('click', event => {
  
  const searchTextElement = document.getElementById("searchText") as HTMLInputElement;

  fetch(`/query?q=${searchTextElement.value}`)
  .then(response => response.json())
  .then(data => {
    const resultsElement = document.getElementById("searchResults");
    
    // empty out old results
    while (resultsElement.firstChild) {
      resultsElement.removeChild(resultsElement.lastChild);
    }

    if (data.torrents) {
      // add in new results
      for (const torrent of data.torrents) {
        addSearchResult(resultsElement, torrent);
      }
    } else {
      addSearchResult(resultsElement, { name: "No Results" });
    }
  });
});

function addSearchResult(resultsElement: HTMLElement, torrent: any, ) {
  const torrentElement = document.createElement("div");
  
  const nameElement = document.createElement("span");
  nameElement.innerText = torrent.name;
  torrentElement.appendChild(nameElement);

  resultsElement.appendChild(torrentElement);
}
