const tracks = [
  { id: 1, title: "Теплее", artist: "Северный ветер", genre: "Инди", duration: "3:42", cover: "cover-sunset" },
  { id: 2, title: "Пока горят окна", artist: "Линии", genre: "Электроника", duration: "4:08", cover: "cover-violet" },
  { id: 3, title: "Медленный свет", artist: "Комната 12", genre: "Альтернатива", duration: "3:16", cover: "cover-blue" },
  { id: 4, title: "После дождя", artist: "Миша Л.", genre: "Хип-хоп", duration: "2:58", cover: "cover-green" },
  { id: 5, title: "Между нами", artist: "Северный ветер", genre: "Инди", duration: "3:31", cover: "cover-violet" },
  { id: 6, title: "Тихий район", artist: "Линии", genre: "Электроника", duration: "4:21", cover: "cover-blue" },
];

const playlists = [
  { id: 1, name: "Ночная смена", trackIds: [1, 2, 3, 4], color: "#b46a57" },
  { id: 2, name: "На повторе", trackIds: [5, 6], color: "#7891a4" },
  { id: 3, name: "Тихие утра", trackIds: [], color: "#849575" },
];

let selectedPlaylistId = 1;
let currentTrackId = null;
let isPlaying = false;
let showingAllTracks = false;

const trackList = document.querySelector("#track-list");
const emptyState = document.querySelector("#empty-state");
const playlistNav = document.querySelector("#playlist-nav");
const trackDialog = document.querySelector("#track-dialog");
const playlistDialog = document.querySelector("#playlist-dialog");

function getSelectedPlaylist() {
  return playlists.find((playlist) => playlist.id === selectedPlaylistId);
}

function formatTrackCount(count) {
  const lastDigit = count % 10;
  const lastTwoDigits = count % 100;
  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return `${count} треков`;
  if (lastDigit === 1) return `${count} трек`;
  if (lastDigit >= 2 && lastDigit <= 4) return `${count} трека`;
  return `${count} треков`;
}

function renderPlaylists() {
  playlistNav.replaceChildren();
  for (const playlist of playlists) {
    const link = document.createElement("a");
    link.className = `playlist-link${playlist.id === selectedPlaylistId ? " selected" : ""}`;
    link.href = "#library";
    link.dataset.playlistId = String(playlist.id);

    const dot = document.createElement("span");
    dot.className = "playlist-dot";
    dot.style.setProperty("--dot", playlist.color);

    const name = document.createElement("span");
    name.textContent = playlist.name;
    link.append(dot, name);
    playlistNav.append(link);
  }
}

function makeTrackRow(track, index) {
  const row = document.createElement("article");
  row.className = "track-row";

  const number = document.createElement("span");
  number.className = "track-index";
  number.textContent = String(index + 1).padStart(2, "0");

  const play = document.createElement("button");
  play.className = `track-play${currentTrackId === track.id && isPlaying ? " playing" : ""}`;
  play.type = "button";
  play.dataset.playTrack = String(track.id);
  play.setAttribute("aria-label", `Слушать: ${track.title}`);
  play.textContent = currentTrackId === track.id && isPlaying ? "Ⅱ" : "▶";

  const copy = document.createElement("div");
  copy.className = "track-copy";
  const title = document.createElement("strong");
  title.textContent = track.title;
  const artist = document.createElement("span");
  artist.textContent = track.artist;
  copy.append(title, artist);

  const genre = document.createElement("span");
  genre.className = "track-genre";
  genre.textContent = track.genre;

  const duration = document.createElement("span");
  duration.className = "track-duration";
  duration.textContent = track.duration;

  const add = document.createElement("button");
  add.className = "track-add";
  add.type = "button";
  add.dataset.addTrack = String(track.id);
  add.setAttribute("aria-label", `Добавить «${track.title}» в плейлист`);
  add.textContent = "+";

  row.append(number, play, copy, genre, duration, add);
  return row;
}

function renderTracks() {
  const playlist = getSelectedPlaylist();
  const query = document.querySelector("#search-input").value.trim().toLocaleLowerCase("ru");
  const playlistTracks = showingAllTracks ? tracks : playlist.trackIds.map((id) => tracks.find((track) => track.id === id)).filter(Boolean);
  const filteredTracks = playlistTracks.filter((track) => `${track.title} ${track.artist} ${track.genre}`.toLocaleLowerCase("ru").includes(query));

  trackList.replaceChildren(...filteredTracks.map(makeTrackRow));
  emptyState.classList.toggle("hidden", filteredTracks.length !== 0);
  document.querySelector("#playlist-title").textContent = playlist.name;
  document.querySelector("#track-section-title").textContent = showingAllTracks ? "Все треки" : "Треки плейлиста";
  document.querySelector("#featured-track-count").textContent = formatTrackCount(playlist.trackIds.length);
  document.querySelector("#sort-button").classList.toggle("hidden", showingAllTracks);
}

function renderSuggestions() {
  const playlist = getSelectedPlaylist();
  const suggestions = tracks.filter((track) => !playlist.trackIds.includes(track.id)).slice(0, 4);
  const grid = document.querySelector("#suggestion-grid");
  grid.replaceChildren();

  for (const track of suggestions) {
    const card = document.createElement("article");
    card.className = "suggestion-card";

    const cover = document.createElement("div");
    cover.className = `suggestion-cover ${track.cover}`;
    const symbol = document.createElement("span");
    symbol.className = "cover-symbol";
    symbol.textContent = "R";
    cover.append(symbol);

    const info = document.createElement("div");
    info.className = "suggestion-info";
    const copy = document.createElement("div");
    const title = document.createElement("strong");
    title.textContent = track.title;
    const artist = document.createElement("span");
    artist.textContent = track.artist;
    copy.append(title, artist);

    const add = document.createElement("button");
    add.className = "suggestion-add";
    add.type = "button";
    add.dataset.addTrack = String(track.id);
    add.setAttribute("aria-label", `Добавить «${track.title}» в плейлист`);
    add.textContent = "+";
    info.append(copy, add);
    card.append(cover, info);
    grid.append(card);
  }

  if (suggestions.length === 0) {
    const note = document.createElement("p");
    note.className = "form-hint";
    note.textContent = "Все доступные треки уже в этом плейлисте.";
    grid.append(note);
  }
}

function render() {
  renderPlaylists();
  renderTracks();
  renderSuggestions();
}

function setCurrentTrack(trackId) {
  const track = tracks.find((item) => item.id === trackId);
  if (!track) return;
  currentTrackId = track.id;
  isPlaying = true;
  document.querySelector("#player-track-name").textContent = track.title;
  document.querySelector("#player-artist-name").textContent = track.artist;
  document.querySelector("#player-cover").className = `mini-cover ${track.cover}`;
  document.querySelector("#duration-time").textContent = track.duration;
  document.querySelector("#elapsed-time").textContent = "0:00";
  document.querySelector("#progress-fill").style.width = "0%";
  document.querySelector("#player-play-button").textContent = "Ⅱ";
  document.querySelector("#player-play-button").setAttribute("aria-label", "Пауза");
  renderTracks();
}

function addTrackToPlaylist(trackId) {
  const playlist = getSelectedPlaylist();
  if (!playlist.trackIds.includes(trackId)) {
    playlist.trackIds.push(trackId);
    showingAllTracks = false;
    render();
  }
}

document.querySelector("#playlist-nav").addEventListener("click", (event) => {
  const link = event.target.closest("[data-playlist-id]");
  if (!link) return;
  event.preventDefault();
  selectedPlaylistId = Number(link.dataset.playlistId);
  showingAllTracks = false;
  render();
});

document.querySelector("#track-list").addEventListener("click", (event) => {
  const playButton = event.target.closest("[data-play-track]");
  if (playButton) {
    const trackId = Number(playButton.dataset.playTrack);
    if (currentTrackId === trackId) {
      isPlaying = !isPlaying;
      document.querySelector("#player-play-button").textContent = isPlaying ? "Ⅱ" : "▶";
      document.querySelector("#player-play-button").setAttribute("aria-label", isPlaying ? "Пауза" : "Воспроизвести");
      renderTracks();
    } else {
      setCurrentTrack(trackId);
    }
    return;
  }

  const addButton = event.target.closest("[data-add-track]");
  if (addButton) addTrackToPlaylist(Number(addButton.dataset.addTrack));
});

document.querySelector("#suggestion-grid").addEventListener("click", (event) => {
  const button = event.target.closest("[data-add-track]");
  if (button) addTrackToPlaylist(Number(button.dataset.addTrack));
});

document.querySelector("#search-input").addEventListener("input", () => {
  showingAllTracks = true;
  renderTracks();
});

document.querySelector("#show-all-button").addEventListener("click", () => {
  showingAllTracks = true;
  renderTracks();
  document.querySelector("#search-input").focus();
});

document.querySelector("#play-playlist-button").addEventListener("click", () => {
  const firstTrack = getSelectedPlaylist().trackIds[0];
  if (firstTrack) setCurrentTrack(firstTrack);
  else trackDialog.showModal();
});

document.querySelector("#player-play-button").addEventListener("click", () => {
  if (currentTrackId === null) {
    const firstTrack = getSelectedPlaylist().trackIds[0];
    if (firstTrack) setCurrentTrack(firstTrack);
    return;
  }
  isPlaying = !isPlaying;
  document.querySelector("#player-play-button").textContent = isPlaying ? "Ⅱ" : "▶";
  document.querySelector("#player-play-button").setAttribute("aria-label", isPlaying ? "Пауза" : "Воспроизвести");
  renderTracks();
});

document.querySelector("#player-like-button").addEventListener("click", (event) => {
  const button = event.currentTarget;
  const liked = button.classList.toggle("liked");
  button.textContent = liked ? "♥" : "♡";
  button.setAttribute("aria-label", liked ? "Убрать из избранного" : "Добавить в избранное");
});

document.querySelector("#add-track-button").addEventListener("click", () => trackDialog.showModal());
document.querySelector("#add-to-playlist-button").addEventListener("click", () => trackDialog.showModal());
document.querySelector("#create-playlist-button").addEventListener("click", () => playlistDialog.showModal());

document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => button.closest("dialog").close());
});

document.querySelector("#track-form").addEventListener("submit", (event) => {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);
  const title = String(formData.get("title")).trim();
  const artist = String(formData.get("artist")).trim();
  const genre = String(formData.get("genre"));
  const newTrack = {
    id: Math.max(0, ...tracks.map((track) => track.id)) + 1,
    title,
    artist,
    genre,
    duration: "0:00",
    cover: ["cover-sunset", "cover-violet", "cover-blue", "cover-green"][tracks.length % 4],
  };
  tracks.unshift(newTrack);
  getSelectedPlaylist().trackIds.unshift(newTrack.id);
  event.currentTarget.reset();
  trackDialog.close();
  showingAllTracks = false;
  render();
});

document.querySelector("#playlist-form").addEventListener("submit", (event) => {
  event.preventDefault();
  const formData = new FormData(event.currentTarget);
  const name = String(formData.get("name")).trim();
  if (!name) return;
  const colors = ["#b46a57", "#7891a4", "#849575", "#9478a9"];
  const newPlaylist = {
    id: Math.max(0, ...playlists.map((playlist) => playlist.id)) + 1,
    name,
    trackIds: [],
    color: colors[playlists.length % colors.length],
  };
  playlists.push(newPlaylist);
  selectedPlaylistId = newPlaylist.id;
  event.currentTarget.reset();
  playlistDialog.close();
  showingAllTracks = false;
  render();
});

document.querySelector("#sort-button").addEventListener("click", () => {
  getSelectedPlaylist().trackIds.reverse();
  render();
});

render();
