/// <reference path="../../node_modules/@types/zepto/index.d.ts" />

import { QueryResponse } from "./searchdata";
import { generateLink } from "./magnet";

$("#searchButton").on("click", _ => {

  $.ajax({
    type: 'GET',
    url: '/query',
    // data to be added to query string:
    data: { q: $("#searchText").val() },
    // type of data we are expecting in return:
    dataType: 'json',
    timeout: 300,
    success: function (data: QueryResponse) {

      const $searchResults = $("#searchResults")
      $searchResults.empty();

      if (data.torrents) {
        // add in new results
        for (const torrent of data.torrents) {
          const $torrentElement = $("<div>")
          $torrentElement.append($("<span>").text(torrent.name));
          $torrentElement.append($(`<a href="${generateLink(torrent.infohashHex)}">🧲</a>`));
          $searchResults.append($torrentElement);
        }
      } else {
        const $torrentElement = $("<div>")
        $torrentElement.append($("<span>").text("No Results"));
        $searchResults.append($torrentElement);
      }
    },
    error: function () {
      const $searchResults = $("#searchResults")
      $searchResults.empty();
      $searchResults.append($("<div>").append($("<span>").text("Error getting results")));
    }
  });

});


