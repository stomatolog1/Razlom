const tracks = [];
let publicPlaylists = [];
let ownedPlaylists = [];
let playlists = [];
let likedTrackIds = new Set();
let repostedTrackIds = new Set();
let randomTrackIds = [];
let currentUser = null;
let selectedPlaylistId = null;
let currentTrackId = null;
let isPlaying = false;
let showingAllTracks = true;
let newestFirst = true;
let authMode = "login";
let profileData = null;
let adminData = null;
let profileRepostCount = 25;
let activeView = "home";
let refreshInProgress = false;

const trackList = document.querySelector("#track-list");
const emptyState = document.querySelector("#empty-state");
const playlistNav = document.querySelector("#playlist-nav");
const trackDialog = document.querySelector("#track-dialog");
const playlistDialog = document.querySelector("#playlist-dialog");
const authDialog = document.querySelector("#auth-dialog");
const apiError = document.querySelector("#api-error");
const trackFormError = document.querySelector("#track-form-error");
const authError = document.querySelector("#auth-error");
const audioPlayer = document.querySelector("#audio-player");
const coverClasses = ["cover-sunset", "cover-violet", "cover-blue", "cover-green"];
const playlistColors = ["#b46a57", "#7891a4", "#849575", "#9478a9"];

function showApiError(message) {
  apiError.textContent = message;
  apiError.classList.remove("hidden");
}

function hideApiError() {
  apiError.textContent = "";
  apiError.classList.add("hidden");
}

async function api(path, options = {}) {
  const response = await fetch(path, { credentials: "same-origin", ...options });
  const text = await response.text();
  let body = null;
  if (text) {
    try {
      body = JSON.parse(text);
    } catch {
      if (response.ok) throw new Error("Сервер вернул некорректный ответ.");
    }
  }
  if (!response.ok) {
    const error = new Error(body?.error || `Ошибка сервера (${response.status}).`);
    error.status = response.status;
    throw error;
  }
  return body;
}

function jsonOptions(method, value) {
  return {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(value),
  };
}

function requireLogin() {
  if (currentUser) return true;
  openAuthDialog("login");
  return false;
}

function formatDuration(seconds) {
  if (!Number.isFinite(seconds) || seconds < 0) return "0:00";
  const minutes = Math.floor(seconds / 60);
  const remainder = Math.floor(seconds % 60);
  return `${minutes}:${String(remainder).padStart(2, "0")}`;
}

function mapTrack(track) {
  return {
    id: Number(track.id),
    title: track.name,
    artist: track.artist,
    genre: track.genre,
    duration: formatDuration(Number(track.time)),
    audioUrl: track.pathToMusic || "",
    uploaderId: Number(track.username),
    cover: coverClasses[Number(track.id) % coverClasses.length],
  };
}

function mapPlaylist(playlist) {
  return {
    id: Number(playlist.id),
    name: playlist.name,
    trackIds: Array.isArray(playlist.id_music) ? playlist.id_music.map(Number) : [],
    ownerId: Number(playlist.id_user),
    private: Boolean(playlist.private),
    color: playlistColors[Number(playlist.id) % playlistColors.length],
  };
}

function asArray(value, label) {
  if (!Array.isArray(value)) throw new Error(`Сервер вернул некорректные данные: ${label}.`);
  return value;
}

function synchronizePlaylists() {
  const byID = new Map();
  for (const playlist of publicPlaylists) byID.set(playlist.id, playlist);
  for (const playlist of ownedPlaylists) byID.set(playlist.id, playlist);
  playlists = [...byID.values()];
  if (!playlists.some((playlist) => playlist.id === selectedPlaylistId)) {
    selectedPlaylistId = ownedPlaylists[0]?.id ?? publicPlaylists[0]?.id ?? null;
  }
}

async function refreshData({ silent = false } = {}) {
  if (refreshInProgress) return;
  refreshInProgress = true;
  try {
    const [musicData, publicPlaylistData] = await Promise.all([
      api("/api/music"),
      api("/api/playlists"),
    ]);
    const nextTracks = asArray(musicData, "music").map(mapTrack);
    const nextPublicPlaylists = asArray(publicPlaylistData, "playlists").map(mapPlaylist);

    let nextOwnedPlaylists = [];
    let nextLikes = new Set();
    let nextReposts = new Set();
    if (currentUser) {
      const [ownedData, likesData, repostsData] = await Promise.all([
        api("/api/my/playlists"),
        api("/api/likes"),
        api("/api/reposts"),
      ]);
      nextOwnedPlaylists = asArray(ownedData, "my playlists").map(mapPlaylist);
      nextLikes = new Set(asArray(likesData, "likes").flatMap((like) => like.id_music.map(Number)));
      nextReposts = new Set(asArray(repostsData, "reposts").flatMap((repost) => repost.id_music.map(Number)));
    }

    tracks.splice(0, tracks.length, ...nextTracks);
    publicPlaylists = nextPublicPlaylists;
    ownedPlaylists = nextOwnedPlaylists;
    likedTrackIds = nextLikes;
    repostedTrackIds = nextReposts;
    randomTrackIds = randomTrackIds.filter((id) => nextTracks.some((track) => track.id === id));
    const remainingRandomTracks = nextTracks.filter((track) => !randomTrackIds.includes(track.id));
    for (let index = remainingRandomTracks.length - 1; index > 0; index -= 1) {
      const randomIndex = Math.floor(Math.random() * (index + 1));
      [remainingRandomTracks[index], remainingRandomTracks[randomIndex]] =
        [remainingRandomTracks[randomIndex], remainingRandomTracks[index]];
    }
    randomTrackIds.push(
      ...remainingRandomTracks.slice(0, Math.max(0, 10 - randomTrackIds.length)).map((track) => track.id),
    );
    synchronizePlaylists();
    hideApiError();
  } catch (error) {
    if (error.status === 401 && currentUser) {
      currentUser = null;
      activeView = "home";
      ownedPlaylists = [];
      likedTrackIds.clear();
      repostedTrackIds.clear();
      synchronizePlaylists();
      updateAccountControls();
      showApiError("Сессия завершилась. Войдите снова, чтобы продолжить.");
    } else if (!silent || !apiError.classList.contains("hidden")) {
      showApiError(`Не удалось обновить данные: ${error.message}`);
    }
  } finally {
    refreshInProgress = false;
    render();
  }
}

function getSelectedPlaylist() {
  return playlists.find((playlist) => playlist.id === selectedPlaylistId) || null;
}

function ownsPlaylist(playlist) {
  return Boolean(currentUser && playlist && playlist.ownerId === currentUser.id);
}

function formatTrackCount(count) {
  const lastDigit = count % 10;
  const lastTwoDigits = count % 100;
  if (lastTwoDigits >= 11 && lastTwoDigits <= 14) return `${count} треков`;
  if (lastDigit === 1) return `${count} трек`;
  if (lastDigit >= 2 && lastDigit <= 4) return `${count} трека`;
  return `${count} треков`;
}

function updateAccountControls() {
  const authButton = document.querySelector("#auth-button");
  const addTrackButton = document.querySelector("#add-track-button");
  const profileNavLink = document.querySelector("#profile-nav-link");

  authButton.textContent = currentUser ? "Выйти" : "Войти";
  authButton.setAttribute("aria-label", currentUser ? "Выйти из аккаунта" : "Войти");
  document.querySelector("#user-name").textContent = currentUser?.username || "Гость";
  document.querySelector("#user-status").textContent = currentUser?.isAdmin ? "Администратор" : currentUser ? "Слушатель" : "Публичный каталог";
  document.querySelector("#user-avatar").textContent = currentUser?.username?.slice(0, 1).toLocaleUpperCase("ru") || "?";
  addTrackButton.classList.toggle("hidden", !currentUser);
  profileNavLink.setAttribute("aria-label", currentUser ? "Открыть мой профиль" : "Войти, чтобы открыть профиль");
  document.querySelector("#admin-nav-link").classList.toggle("hidden", !currentUser?.isAdmin);
  document.querySelector("#sidebar-profile-button").setAttribute(
    "aria-label",
    currentUser ? `Открыть профиль ${currentUser.username}` : "Войти, чтобы открыть профиль",
  );
}

function renderPlaylistNavigation() {
  playlistNav.replaceChildren();
  for (const playlist of ownedPlaylists) {
    const link = document.createElement("a");
    link.className = `playlist-link${playlist.id === selectedPlaylistId ? " selected" : ""}`;
    link.href = "#library";
    link.dataset.playlistId = String(playlist.id);

    const dot = document.createElement("span");
    dot.className = "playlist-dot";
    dot.style.setProperty("--dot", playlist.color);

    const name = document.createElement("span");
    name.textContent = playlist.name;
    if (playlist.private) {
      const privacy = document.createElement("span");
      privacy.className = "playlist-privacy";
      privacy.textContent = "●";
      privacy.title = "Приватный плейлист";
      link.append(dot, name, privacy);
    } else {
      link.append(dot, name);
    }
    playlistNav.append(link);
  }
}

function makePlaylistCard(playlist) {
  const card = document.createElement("article");
  card.className = "public-playlist-card";
  card.style.setProperty("--playlist-color", playlist.color);

  const cover = document.createElement("div");
  cover.className = "public-playlist-cover";
  cover.textContent = "R";

  const info = document.createElement("div");
  info.className = "public-playlist-info";
  const name = document.createElement("strong");
  name.textContent = playlist.name;
  const detail = document.createElement("span");
  detail.textContent = `${formatTrackCount(playlist.trackIds.length)} · ${playlist.private ? "Приватный" : "Публичный"}`;

  const open = document.createElement("button");
  open.className = "outline-button playlist-open-button";
  open.type = "button";
  open.dataset.selectPlaylist = String(playlist.id);
  open.textContent = "Открыть";
  info.append(name, detail, open);
  card.append(cover, info);

  if (currentUser && (playlist.ownerId === currentUser.id || currentUser.isAdmin)) {
    const remove = document.createElement("button");
    remove.className = "delete-button playlist-delete-button";
    remove.type = "button";
    remove.dataset.deletePlaylist = String(playlist.id);
    remove.textContent = "Удалить";
    remove.setAttribute("aria-label", `Удалить плейлист «${playlist.name}»`);
    card.append(remove);
  } else {
    const owner = document.createElement("button");
    owner.className = "playlist-owner";
    owner.type = "button";
    owner.dataset.openProfile = String(playlist.ownerId);
    owner.textContent = `Профиль пользователя #${playlist.ownerId}`;
    card.append(owner);
  }
  return card;
}

function renderPublicPlaylists() {
  const grid = document.querySelector("#public-playlist-grid");
  grid.replaceChildren();
  for (const playlist of publicPlaylists) grid.append(makePlaylistCard(playlist));
  if (publicPlaylists.length === 0) {
    const note = document.createElement("p");
    note.className = "form-hint";
    note.textContent = "Публичных плейлистов пока нет.";
    grid.append(note);
  }
}

function makeRandomTrackCard(track) {
  const card = document.createElement("article");
  card.className = "random-track-card";

  const cover = document.createElement("div");
  cover.className = `random-track-cover ${track.cover}`;
  const play = document.createElement("button");
  play.className = "random-play-button";
  play.type = "button";
  play.dataset.playTrack = String(track.id);
  play.textContent = currentTrackId === track.id && isPlaying ? "Ⅱ" : "▶";
  play.setAttribute("aria-label", `Слушать: ${track.title}`);
  cover.append(play);

  const info = document.createElement("div");
  info.className = "random-track-info";
  const title = document.createElement("strong");
  title.textContent = track.title;
  const artist = document.createElement("span");
  artist.textContent = track.artist;
  info.append(title, artist);

  const like = document.createElement("button");
  like.className = `track-action${likedTrackIds.has(track.id) ? " liked" : ""}`;
  like.type = "button";
  like.dataset.likeTrack = String(track.id);
  like.textContent = likedTrackIds.has(track.id) ? "♥" : "♡";
  like.setAttribute("aria-label", likedTrackIds.has(track.id) ? "Убрать из избранного" : "Добавить в избранное");
  like.disabled = !currentUser;
  card.append(cover, info, like);
  return card;
}

function renderHome() {
  const liked = tracks.filter((track) => likedTrackIds.has(track.id));
  const likedList = document.querySelector("#liked-track-list");
  likedList.replaceChildren(...liked.map(makeTrackRow));
  const likedEmpty = document.querySelector("#liked-empty");
  likedEmpty.classList.toggle("hidden", liked.length > 0);
  likedEmpty.textContent = currentUser ? "Пока нет понравившихся песен." : "Войди, чтобы увидеть понравившиеся песни.";

  const randomTracks = randomTrackIds
    .map((id) => tracks.find((track) => track.id === id))
    .filter(Boolean);
  const carousel = document.querySelector("#random-track-carousel");
  carousel.replaceChildren(...randomTracks.map(makeRandomTrackCard));
  if (randomTracks.length === 0) {
    const note = document.createElement("p");
    note.className = "form-hint";
    note.textContent = "Пока нет треков для рекомендаций.";
    carousel.append(note);
  }
  updateRandomEndNote();
}

function updateRandomEndNote() {
  const carousel = document.querySelector("#random-track-carousel");
  const note = document.querySelector("#random-end-note");
  const hasTracks = randomTrackIds.length > 0;
  const atEnd = carousel.scrollLeft + carousel.clientWidth >= carousel.scrollWidth - 8;
  note.classList.toggle("hidden", !hasTracks || !atEnd);
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

  const owner = document.createElement("button");
  owner.className = "track-owner-button";
  owner.type = "button";
  owner.dataset.openProfile = String(track.uploaderId);
  owner.textContent = "Профиль";

  const like = document.createElement("button");
  like.className = `track-action${likedTrackIds.has(track.id) ? " liked" : ""}`;
  like.type = "button";
  like.dataset.likeTrack = String(track.id);
  like.setAttribute("aria-label", likedTrackIds.has(track.id) ? "Убрать из избранного" : "Добавить в избранное");
  like.textContent = likedTrackIds.has(track.id) ? "♥" : "♡";

  const repost = document.createElement("button");
  repost.className = `track-action${repostedTrackIds.has(track.id) ? " reposted" : ""}`;
  repost.type = "button";
  repost.dataset.repostTrack = String(track.id);
  repost.disabled = repostedTrackIds.has(track.id);
  repost.setAttribute("aria-label", repost.disabled ? "Уже опубликовано в профиле" : "Репостнуть в профиль");
  repost.textContent = repost.disabled ? "↗✓" : "↗";

  const playlist = getSelectedPlaylist();
  const add = document.createElement("button");
  add.className = "track-add";
  add.type = "button";
  add.dataset.addTrack = String(track.id);
  add.hidden = !ownsPlaylist(playlist);
  add.disabled = Boolean(playlist?.trackIds.includes(track.id));
  add.setAttribute("aria-label", `Добавить «${track.title}» в плейлист`);
  add.textContent = add.disabled ? "✓" : "+";

  const remove = document.createElement("button");
  remove.className = "track-delete";
  remove.type = "button";
  remove.dataset.deleteTrack = String(track.id);
  remove.textContent = "×";
  remove.setAttribute("aria-label", `Удалить трек «${track.title}»`);
  remove.hidden = !currentUser || (!currentUser.isAdmin && currentUser.id !== track.uploaderId);

  row.append(number, play, copy, genre, duration, owner, like, repost, add, remove);
  return row;
}

function getVisibleTracks() {
  const playlist = getSelectedPlaylist();
  let visibleTracks = showingAllTracks || !playlist
    ? [...tracks]
    : playlist.trackIds.map((id) => tracks.find((track) => track.id === id)).filter(Boolean);
  if (!newestFirst) visibleTracks.reverse();
  const query = document.querySelector("#search-input").value.trim().toLocaleLowerCase("ru");
  if (query) {
    visibleTracks = visibleTracks.filter((track) =>
      `${track.title} ${track.artist} ${track.genre}`.toLocaleLowerCase("ru").includes(query));
  }
  return visibleTracks;
}

function renderTracks() {
  const playlist = getSelectedPlaylist();
  const visibleTracks = getVisibleTracks();
  trackList.replaceChildren(...visibleTracks.map(makeTrackRow));
  emptyState.classList.toggle("hidden", visibleTracks.length !== 0);
  emptyState.textContent = tracks.length === 0
    ? "Каталог пока пуст. Войди и добавь первый трек."
    : "В этой подборке пока нет треков.";
  document.querySelector("#playlist-title").textContent = playlist?.name || "Музыка сообщества";
  document.querySelector("#track-section-title").textContent = showingAllTracks || !playlist
    ? "Все треки"
    : "Треки плейлиста";
  document.querySelector("#featured-track-count").textContent = formatTrackCount(
    showingAllTracks || !playlist ? tracks.length : playlist.trackIds.length,
  );
  const sortButton = document.querySelector("#sort-button");
  sortButton.classList.toggle("hidden", visibleTracks.length < 2);
  sortButton.firstChild.textContent = newestFirst ? "Сначала новые " : "Сначала старые ";
  document.querySelector("#add-to-playlist-button").disabled = !ownsPlaylist(playlist);
}

function renderSuggestions() {
  const playlist = getSelectedPlaylist();
  const suggestions = (playlist
    ? tracks.filter((track) => !playlist.trackIds.includes(track.id))
    : tracks).slice(0, 4);
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
    add.disabled = !ownsPlaylist(playlist) || playlist.trackIds.includes(track.id);
    add.setAttribute("aria-label", `Добавить «${track.title}» в плейлист`);
    add.textContent = add.disabled ? "✓" : "+";
    info.append(copy, add);
    card.append(cover, info);
    grid.append(card);
  }
  if (suggestions.length === 0) {
    const note = document.createElement("p");
    note.className = "form-hint";
    note.textContent = tracks.length === 0
      ? "Когда появятся треки, они будут доступны всем слушателям."
      : "Все доступные треки уже в этом плейлисте.";
    grid.append(note);
  }
}

function renderProfile() {
  if (activeView !== "profile" || !profileData) return;

  document.querySelector("#profile-title").textContent = profileData.user.username;
  const playlistGrid = document.querySelector("#profile-playlists");
  playlistGrid.replaceChildren();
  const largestPlaylists = [...profileData.playlists]
    .sort((left, right) => right.trackIds.length - left.trackIds.length)
    .slice(0, 5);
  for (const playlist of largestPlaylists) playlistGrid.append(makePlaylistCard(playlist));
  if (profileData.playlists.length === 0) {
    const note = document.createElement("p");
    note.className = "form-hint";
    note.textContent = "У пользователя пока нет плейлистов.";
    playlistGrid.append(note);
  }

  const profileTrackIDs = [...new Set(
    profileData.reposts.flatMap((repost) => repost.id_music.map(Number)),
  )].reverse();
  const profileTracks = profileTrackIDs
    .map((id) => tracks.find((track) => track.id === id))
    .filter(Boolean);
  const repostList = document.querySelector("#profile-reposts");
  const visibleReposts = profileTracks.slice(0, profileRepostCount);
  repostList.replaceChildren(...visibleReposts.map(makeRepostCard));
  document.querySelector("#profile-empty").classList.toggle("hidden", profileTracks.length !== 0);
  document.querySelector("#profile-load-more").classList.toggle("hidden", profileRepostCount >= profileTracks.length);

  const uploadedTracks = (profileData.uploadedTracks || [])
    .map(mapTrack)
    .filter((track) => tracks.some((item) => item.id === track.id));
  const uploadedList = document.querySelector("#profile-uploaded-tracks");
  uploadedList.replaceChildren(...uploadedTracks.map(makeTrackRow));
  document.querySelector("#profile-tracks-empty").classList.toggle("hidden", uploadedTracks.length !== 0);
}

function renderAdmin() {
  if (activeView !== "admin" || !adminData) return;
  const usersList = document.querySelector("#admin-users-list");
  usersList.replaceChildren();
  for (const user of adminData.users) {
    const row = document.createElement("article");
    row.className = "admin-row";
    const label = document.createElement("div");
    label.className = "admin-row-copy";
    const name = document.createElement("strong");
    name.textContent = user.username;
    const detail = document.createElement("span");
    detail.textContent = `@${user.login} · ID ${user.id}${user.isAdmin ? " · администратор" : ""}`;
    label.append(name, detail);
    row.append(label);
    if (user.id !== currentUser.id) {
      const remove = makeDeleteButton("Пользователя", user.username, "deleteUser", user.id);
      row.append(remove);
    }
    usersList.append(row);
  }

  const musicList = document.querySelector("#admin-music-list");
  musicList.replaceChildren();
  for (const rawTrack of adminData.music) {
    const track = mapTrack(rawTrack);
    const row = document.createElement("article");
    row.className = "admin-row";
    const label = document.createElement("div");
    label.className = "admin-row-copy";
    const name = document.createElement("strong");
    name.textContent = track.title;
    const detail = document.createElement("span");
    detail.textContent = `${track.artist} · пользователь #${track.uploaderId}`;
    label.append(name, detail);
    const remove = makeDeleteButton("Трек", track.title, "deleteTrack", track.id);
    row.append(label, remove);
    musicList.append(row);
  }

  const playlistsList = document.querySelector("#admin-playlists-list");
  playlistsList.replaceChildren();
  for (const rawPlaylist of adminData.playlists) {
    const playlist = mapPlaylist(rawPlaylist);
    const row = document.createElement("article");
    row.className = "admin-row";
    const label = document.createElement("div");
    label.className = "admin-row-copy";
    const name = document.createElement("strong");
    name.textContent = playlist.name;
    const detail = document.createElement("span");
    detail.textContent = `Пользователь #${playlist.ownerId} · ${playlist.private ? "приватный" : "открытый"}`;
    label.append(name, detail);
    const remove = makeDeleteButton("Плейлист", playlist.name, "deletePlaylist", playlist.id);
    row.append(label, remove);
    playlistsList.append(row);
  }
}

function makeDeleteButton(kind, name, dataAttribute, id) {
  const button = document.createElement("button");
  button.className = "delete-button";
  button.type = "button";
  button.dataset[dataAttribute] = String(id);
  button.textContent = "Удалить";
  button.setAttribute("aria-label", `Удалить ${kind.toLocaleLowerCase("ru")} «${name}»`);
  return button;
}

function makeRepostCard(track, index) {
  const card = document.createElement("article");
  card.className = "repost-card";
  const cover = document.createElement("div");
  cover.className = `repost-cover ${track.cover}`;
  const number = document.createElement("span");
  number.className = "repost-number";
  number.textContent = String(index + 1).padStart(2, "0");
  const play = document.createElement("button");
  play.className = "random-play-button";
  play.type = "button";
  play.dataset.playTrack = String(track.id);
  play.textContent = currentTrackId === track.id && isPlaying ? "Ⅱ" : "▶";
  play.setAttribute("aria-label", `Слушать: ${track.title}`);
  cover.append(number, play);

  const title = document.createElement("strong");
  title.textContent = track.title;
  const artist = document.createElement("span");
  artist.textContent = track.artist;
  card.append(cover, title, artist);
  return card;
}

function renderPlayer() {
  const track = tracks.find((item) => item.id === currentTrackId);
  const likeButton = document.querySelector("#player-like-button");
  if (!track) {
    likeButton.classList.remove("liked");
    likeButton.disabled = !currentUser;
    return;
  }
  document.querySelector("#player-track-name").textContent = track.title;
  document.querySelector("#player-artist-name").textContent = track.artist;
  document.querySelector("#player-cover").className = `mini-cover ${track.cover}`;
  document.querySelector("#duration-time").textContent = track.duration;
  document.querySelector("#player-play-button").textContent = isPlaying ? "Ⅱ" : "▶";
  document.querySelector("#player-play-button").setAttribute("aria-label", isPlaying ? "Пауза" : "Воспроизвести");
  const liked = likedTrackIds.has(track.id);
  likeButton.classList.toggle("liked", liked);
  likeButton.textContent = liked ? "♥" : "♡";
  likeButton.setAttribute("aria-label", liked ? "Убрать из избранного" : "Добавить в избранное");
  likeButton.disabled = !currentUser;
}

function render() {
  updateAccountControls();
  renderPlaylistNavigation();
  renderPublicPlaylists();
  renderHome();
  renderTracks();
  renderSuggestions();
  renderPlayer();
  document.querySelector("#home-view").classList.toggle("hidden", activeView !== "home");
  document.querySelector("#library-view").classList.toggle("hidden", activeView !== "library");
  document.querySelector("#profile-view").classList.toggle("hidden", activeView !== "profile");
  document.querySelector("#admin-view").classList.toggle("hidden", activeView !== "admin");
  document.querySelectorAll(".primary-nav .nav-link").forEach((link) => {
    const activeLink = activeView === "library"
      ? link.id === "discover-nav-link"
      : link.id === `${activeView}-nav-link`;
    link.classList.toggle("active", activeLink);
  });
  updateRandomEndNote();
  renderProfile();
  renderAdmin();
}

function openHome() {
  activeView = "home";
  document.querySelector("#home").scrollIntoView({ behavior: "smooth" });
  render();
}

async function openAdmin() {
  if (!currentUser?.isAdmin) return;
  activeView = "admin";
  adminData = null;
  render();
  const error = document.querySelector("#admin-error");
  error.classList.add("hidden");
  try {
    const [users, music, playlists] = await Promise.all([
      api("/api/admin/users"),
      api("/api/admin/music"),
      api("/api/admin/playlists"),
    ]);
    adminData = {
      users: asArray(users, "admin users"),
      music: asArray(music, "admin music"),
      playlists: asArray(playlists, "admin playlists"),
    };
    render();
  } catch (requestError) {
    error.textContent = `Не удалось загрузить админ-панель: ${requestError.message}`;
    error.classList.remove("hidden");
  }
}

function openAuthDialog(mode = "login") {
  authMode = mode;
  const register = authMode === "register";
  document.querySelector("#auth-title").textContent = register ? "Создать аккаунт" : "Войти";
  document.querySelector("#auth-submit").textContent = register ? "Зарегистрироваться" : "Войти";
  document.querySelector("#auth-mode-button").textContent = register ? "Уже есть аккаунт?" : "Создать аккаунт";
  const usernameField = document.querySelector("#username-field");
  usernameField.classList.toggle("hidden", !register);
  usernameField.querySelector("input").required = register;
  document.querySelector('#auth-form [name="password"]').autocomplete = register ? "new-password" : "current-password";
  authError.classList.add("hidden");
  if (!authDialog.open) authDialog.showModal();
}

function setCurrentTrack(trackId) {
  const track = tracks.find((item) => item.id === trackId);
  if (!track) return;
  if (!track.audioUrl) {
    showApiError("У этого трека нет доступного аудиофайла.");
    return;
  }
  currentTrackId = track.id;
  audioPlayer.src = track.audioUrl;
  audioPlayer.load();
  isPlaying = true;
  audioPlayer.play().catch((error) => {
    isPlaying = false;
    showApiError(`Не удалось воспроизвести трек: ${error.message}`);
    render();
  });
  hideApiError();
  render();
}

function togglePlayback() {
  if (currentTrackId === null) {
    const firstTrack = getVisibleTracks()[0] || tracks[0];
    if (firstTrack) setCurrentTrack(firstTrack.id);
    return;
  }
  if (audioPlayer.paused) {
    audioPlayer.play().then(() => {
      isPlaying = true;
      render();
    }).catch((error) => showApiError(`Не удалось воспроизвести трек: ${error.message}`));
  } else {
    audioPlayer.pause();
    isPlaying = false;
    render();
  }
}

function skipTrack(direction) {
  const queue = getVisibleTracks();
  if (queue.length === 0) return;
  const index = queue.findIndex((track) => track.id === currentTrackId);
  const nextIndex = index < 0
    ? 0
    : (index + direction + queue.length) % queue.length;
  setCurrentTrack(queue[nextIndex].id);
}

async function toggleLike(trackId) {
  if (!requireLogin()) return;
  try {
    if (likedTrackIds.has(trackId)) {
      await api(`/api/likes/${trackId}`, { method: "DELETE" });
      likedTrackIds.delete(trackId);
    } else {
      await api("/api/likes", jsonOptions("POST", { id_music: trackId }));
      likedTrackIds.add(trackId);
    }
    hideApiError();
    render();
  } catch (error) {
    showApiError(`Не удалось обновить избранное: ${error.message}`);
  }
}

async function repostTrack(trackId) {
  if (!requireLogin()) return;
  if (repostedTrackIds.has(trackId)) return;
  try {
    await api("/api/reposts", jsonOptions("POST", { id_music: trackId }));
    repostedTrackIds.add(trackId);
    if (profileData?.user.id === currentUser.id) {
      profileData = await fetchProfile(currentUser.id);
    }
    hideApiError();
    render();
  } catch (error) {
    showApiError(`Не удалось сделать репост: ${error.message}`);
  }
}

async function addTrackToPlaylist(trackId) {
  const playlist = getSelectedPlaylist();
  if (!requireLogin()) return;
  if (!ownsPlaylist(playlist)) {
    showApiError("Выбери свой плейлист, чтобы добавить в него трек.");
    return;
  }
  try {
    const updated = mapPlaylist(await api(
      `/api/playlists/${playlist.id}/tracks`,
      jsonOptions("POST", { music_id: trackId }),
    ));
    ownedPlaylists = ownedPlaylists.map((item) => item.id === updated.id ? updated : item);
    if (!updated.private) {
      publicPlaylists = publicPlaylists.map((item) => item.id === updated.id ? updated : item);
    }
    synchronizePlaylists();
    hideApiError();
    render();
  } catch (error) {
    showApiError(`Не удалось добавить трек в плейлист: ${error.message}`);
  }
}

async function fetchProfile(userId) {
  const user = await api(`/api/users/${userId}`);
  const [playlistData, repostData, uploadedTracks] = await Promise.all([
    currentUser?.id === user.id
      ? api("/api/my/playlists")
      : api(`/api/users/${userId}/playlists`),
    api(`/api/users/${userId}/reposts`),
    api(`/api/users/${userId}/music`),
  ]);
  return {
    user,
    playlists: asArray(playlistData, "profile playlists").map(mapPlaylist),
    reposts: asArray(repostData, "profile reposts"),
    uploadedTracks: asArray(uploadedTracks, "profile music"),
  };
}

async function openProfile(userId) {
  activeView = "profile";
  profileRepostCount = 25;
  profileData = null;
  render();
  try {
    profileData = await fetchProfile(userId);
    render();
  } catch (error) {
    activeView = "home";
    showApiError(`Не удалось открыть профиль: ${error.message}`);
    render();
  }
}

async function handleAuthentication(event) {
  event.preventDefault();
  const form = event.currentTarget;
  const submitButton = document.querySelector("#auth-submit");
  const data = new FormData(form);
  const payload = {
    login: String(data.get("login")).trim(),
    password: String(data.get("password")),
  };
  let endpoint = "/api/sessions";
  if (authMode === "register") {
    endpoint = "/api/users";
    payload.username = String(data.get("username")).trim();
  }
  submitButton.disabled = true;
  authError.classList.add("hidden");
  try {
    currentUser = await api(endpoint, jsonOptions("POST", payload));
    form.reset();
    authDialog.close();
    await refreshData();
  } catch (error) {
    authError.textContent = `Не удалось выполнить вход: ${error.message}`;
    authError.classList.remove("hidden");
  } finally {
    submitButton.disabled = false;
  }
}

async function logout() {
  try {
    await api("/api/session", { method: "DELETE" });
    currentUser = null;
    ownedPlaylists = [];
    likedTrackIds.clear();
    repostedTrackIds.clear();
    profileData = null;
    adminData = null;
    activeView = "home";
    synchronizePlaylists();
    hideApiError();
    render();
  } catch (error) {
    showApiError(`Не удалось выйти: ${error.message}`);
  }
}

function openPlaylist(playlistId) {
  const playlist = playlists.find((item) => item.id === playlistId);
  if (!playlist) return;
  selectedPlaylistId = playlist.id;
  showingAllTracks = false;
  activeView = "library";
  render();
}

async function deleteTrack(trackId) {
  const track = tracks.find((item) => item.id === trackId);
  if (!track || !confirm(`Удалить трек «${track.title}»?`)) return;
  try {
    await api(`/api/music/${trackId}`, { method: "DELETE" });
    if (currentTrackId === trackId) {
      audioPlayer.pause();
      audioPlayer.removeAttribute("src");
      currentTrackId = null;
      isPlaying = false;
    }
    if (profileData) profileData = await fetchProfile(profileData.user.id);
    await refreshData();
    if (activeView === "admin") await openAdmin();
  } catch (error) {
    showApiError(`Не удалось удалить трек: ${error.message}`);
  }
}

async function deletePlaylist(playlistId) {
  const playlist = playlists.find((item) => item.id === playlistId)
    || profileData?.playlists.find((item) => item.id === playlistId)
    || adminData?.playlists.map(mapPlaylist).find((item) => item.id === playlistId);
  if (!playlist || !confirm(`Удалить плейлист «${playlist.name}»?`)) return;
  try {
    await api(`/api/playlists/${playlistId}`, { method: "DELETE" });
    if (activeView === "admin") {
      await refreshData();
      await openAdmin();
    } else {
      await refreshData();
      if (profileData) profileData = await fetchProfile(profileData.user.id);
      render();
    }
  } catch (error) {
    showApiError(`Не удалось удалить плейлист: ${error.message}`);
  }
}

async function deleteUser(userId) {
  const user = adminData?.users.find((item) => item.id === userId);
  if (!user || !confirm(`Удалить пользователя «${user.username}» вместе с его треками, плейлистами и репостами?`)) return;
  try {
    await api(`/api/admin/users/${userId}`, { method: "DELETE" });
    await openAdmin();
    await refreshData();
  } catch (error) {
    const adminError = document.querySelector("#admin-error");
    adminError.textContent = `Не удалось удалить пользователя: ${error.message}`;
    adminError.classList.remove("hidden");
  }
}

document.querySelector("#playlist-nav").addEventListener("click", (event) => {
  const link = event.target.closest("[data-playlist-id]");
  if (!link) return;
  event.preventDefault();
  openPlaylist(Number(link.dataset.playlistId));
});

document.querySelector("#public-playlist-grid").addEventListener("click", (event) => {
  const openButton = event.target.closest("[data-select-playlist]");
  if (openButton) {
    openPlaylist(Number(openButton.dataset.selectPlaylist));
    return;
  }
  const profileButton = event.target.closest("[data-open-profile]");
  if (profileButton) openProfile(Number(profileButton.dataset.openProfile));
});

document.querySelector("#profile-playlists").addEventListener("click", (event) => {
  const openButton = event.target.closest("[data-select-playlist]");
  if (openButton) openPlaylist(Number(openButton.dataset.selectPlaylist));
  const profileButton = event.target.closest("[data-open-profile]");
  if (profileButton) openProfile(Number(profileButton.dataset.openProfile));
});

async function handleTrackAction(event) {
  const playButton = event.target.closest("[data-play-track]");
  if (playButton) {
    const trackId = Number(playButton.dataset.playTrack);
    if (currentTrackId === trackId) togglePlayback();
    else setCurrentTrack(trackId);
    return;
  }
  const likeButton = event.target.closest("[data-like-track]");
  if (likeButton) {
    await toggleLike(Number(likeButton.dataset.likeTrack));
    return;
  }
  const repostButton = event.target.closest("[data-repost-track]");
  if (repostButton) {
    await repostTrack(Number(repostButton.dataset.repostTrack));
    return;
  }
  const addButton = event.target.closest("[data-add-track]");
  if (addButton) await addTrackToPlaylist(Number(addButton.dataset.addTrack));
  const deleteButton = event.target.closest("[data-delete-track]");
  if (deleteButton) await deleteTrack(Number(deleteButton.dataset.deleteTrack));
  const profileButton = event.target.closest("[data-open-profile]");
  if (profileButton) openProfile(Number(profileButton.dataset.openProfile));
}

trackList.addEventListener("click", handleTrackAction);
document.querySelector("#profile-reposts").addEventListener("click", handleTrackAction);
document.querySelector("#profile-uploaded-tracks").addEventListener("click", handleTrackAction);
document.querySelector("#liked-track-list").addEventListener("click", handleTrackAction);
document.querySelector("#random-track-carousel").addEventListener("click", handleTrackAction);
document.querySelector("#suggestion-grid").addEventListener("click", async (event) => {
  const button = event.target.closest("[data-add-track]");
  if (button) await addTrackToPlaylist(Number(button.dataset.addTrack));
});

document.addEventListener("click", async (event) => {
  const deletePlaylistButton = event.target.closest("[data-delete-playlist]");
  if (deletePlaylistButton) {
    await deletePlaylist(Number(deletePlaylistButton.dataset.deletePlaylist));
    return;
  }
  const deleteUserButton = event.target.closest("[data-delete-user]");
  if (deleteUserButton) {
    await deleteUser(Number(deleteUserButton.dataset.deleteUser));
    return;
  }
  const deleteTrackButton = event.target.closest("[data-delete-track]");
  if (deleteTrackButton && !deleteTrackButton.closest(".track-row")) {
    await deleteTrack(Number(deleteTrackButton.dataset.deleteTrack));
    return;
  }
  const carouselButton = event.target.closest("[data-scroll-carousel]");
  if (carouselButton) {
    const carousel = document.getElementById(carouselButton.dataset.scrollCarousel);
    if (carousel) {
      carousel.scrollBy({
        left: Number(carouselButton.dataset.direction) * carousel.clientWidth * 0.8,
        behavior: "smooth",
      });
    }
  }
});

document.querySelector("#random-track-carousel").addEventListener("scroll", updateRandomEndNote, { passive: true });
document.querySelector("#profile-load-more").addEventListener("click", () => {
  profileRepostCount += 25;
  renderProfile();
});

document.querySelector("#search-input").addEventListener("input", () => {
  showingAllTracks = true;
  if (document.querySelector("#search-input").value.trim()) activeView = "library";
  render();
});

document.querySelector("#show-all-button").addEventListener("click", () => {
  showingAllTracks = true;
  activeView = "library";
  render();
  document.querySelector("#search-input").focus();
});

document.querySelector("#play-playlist-button").addEventListener("click", () => {
  const playlist = getSelectedPlaylist();
  const firstTrackId = playlist?.trackIds[0] ?? getVisibleTracks()[0]?.id;
  if (firstTrackId) setCurrentTrack(firstTrackId);
  else showApiError("В этой подборке пока нет треков.");
});

document.querySelector("#add-to-playlist-button").addEventListener("click", () => {
  if (!requireLogin()) return;
  if (!ownsPlaylist(getSelectedPlaylist())) {
    showApiError("Сначала выбери свой плейлист в боковом меню.");
    return;
  }
  if (currentTrackId === null) {
    showApiError("Сначала выбери трек, который нужно добавить.");
    return;
  }
  addTrackToPlaylist(currentTrackId);
});

document.querySelector("#player-play-button").addEventListener("click", togglePlayback);
document.querySelector("#player-like-button").addEventListener("click", () => {
  if (currentTrackId !== null) toggleLike(currentTrackId);
});
document.querySelector("#previous-track-button").addEventListener("click", () => skipTrack(-1));
document.querySelector("#next-track-button").addEventListener("click", () => skipTrack(1));
document.querySelector("#seek-input").addEventListener("input", (event) => {
  if (Number.isFinite(audioPlayer.duration)) {
    audioPlayer.currentTime = audioPlayer.duration * Number(event.currentTarget.value) / 100;
  }
});
document.querySelector("#volume-input").addEventListener("input", (event) => {
  audioPlayer.volume = Number(event.currentTarget.value) / 100;
});

audioPlayer.addEventListener("loadedmetadata", () => {
  const track = tracks.find((item) => item.id === currentTrackId);
  if (track && Number.isFinite(audioPlayer.duration)) {
    track.duration = formatDuration(audioPlayer.duration);
    document.querySelector("#duration-time").textContent = track.duration;
    renderTracks();
  }
});
audioPlayer.addEventListener("timeupdate", () => {
  if (!Number.isFinite(audioPlayer.duration) || audioPlayer.duration <= 0) return;
  document.querySelector("#elapsed-time").textContent = formatDuration(audioPlayer.currentTime);
  document.querySelector("#seek-input").value = String(audioPlayer.currentTime / audioPlayer.duration * 100);
});
audioPlayer.addEventListener("ended", () => {
  if (getVisibleTracks().length > 1) skipTrack(1);
  else {
    isPlaying = false;
    render();
  }
});
audioPlayer.addEventListener("error", () => {
  if (currentTrackId !== null) showApiError("Не удалось загрузить аудиофайл этого трека.");
});

document.querySelector("#add-track-button").addEventListener("click", () => {
  if (requireLogin()) trackDialog.showModal();
});
document.querySelector("#create-playlist-button").addEventListener("click", () => {
  if (requireLogin()) playlistDialog.showModal();
});

document.querySelector("#auth-button").addEventListener("click", () => {
  if (currentUser) logout();
  else openAuthDialog("login");
});
document.querySelector("#sidebar-profile-button").addEventListener("click", () => {
  if (currentUser) openProfile(currentUser.id);
  else openAuthDialog("login");
});
document.querySelector("#auth-mode-button").addEventListener("click", () => {
  openAuthDialog(authMode === "login" ? "register" : "login");
});
document.querySelector("#auth-form").addEventListener("submit", handleAuthentication);

document.querySelector("#playlist-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!requireLogin()) return;
  const form = event.currentTarget;
  const submitButton = form.querySelector('[type="submit"]');
  const formData = new FormData(form);
  submitButton.disabled = true;
  const formError = document.querySelector("#playlist-form-error");
  formError.classList.add("hidden");
  try {
    const playlist = mapPlaylist(await api("/api/playlists", jsonOptions("POST", {
      name: String(formData.get("name")).trim(),
      private: formData.get("private") === "on",
    })));
    ownedPlaylists.unshift(playlist);
    if (!playlist.private) publicPlaylists.unshift(playlist);
    synchronizePlaylists();
    selectedPlaylistId = playlist.id;
    showingAllTracks = false;
    activeView = "library";
    form.reset();
    playlistDialog.close();
    hideApiError();
    render();
  } catch (error) {
    formError.textContent = `Не удалось создать плейлист: ${error.message}`;
    formError.classList.remove("hidden");
  } finally {
    submitButton.disabled = false;
  }
});

document.querySelector("#track-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (!requireLogin()) return;
  const form = event.currentTarget;
  const submitButton = form.querySelector('[type="submit"]');
  submitButton.disabled = true;
  trackFormError.classList.add("hidden");
  try {
    const response = await api("/api/music", {
      method: "POST",
      body: new FormData(form),
    });
    tracks.unshift(mapTrack(response));
    if (randomTrackIds.length < 10) randomTrackIds.push(Number(response.id));
    showingAllTracks = true;
    form.reset();
    trackDialog.close();
    hideApiError();
    render();
  } catch (error) {
    trackFormError.textContent = `Не удалось добавить трек: ${error.message}`;
    trackFormError.classList.remove("hidden");
  } finally {
    submitButton.disabled = false;
  }
});

document.querySelector("#profile-nav-link").addEventListener("click", (event) => {
  event.preventDefault();
  if (currentUser) openProfile(currentUser.id);
  else openAuthDialog("login");
});
document.querySelector("#home-nav-link").addEventListener("click", (event) => {
  event.preventDefault();
  openHome();
});
document.querySelector("#discover-nav-link").addEventListener("click", (event) => {
  event.preventDefault();
  showingAllTracks = true;
  activeView = "library";
  render();
  document.querySelector("#search-input").focus();
});
document.querySelector("#admin-nav-link").addEventListener("click", (event) => {
  event.preventDefault();
  openAdmin();
});

document.querySelector("#sort-button").addEventListener("click", () => {
  newestFirst = !newestFirst;
  renderTracks();
});

document.querySelectorAll("[data-close-dialog]").forEach((button) => {
  button.addEventListener("click", () => button.closest("dialog").close());
});

document.addEventListener("keydown", (event) => {
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "k") {
    event.preventDefault();
    document.querySelector("#search-input").focus();
  }
});

async function initialize() {
  try {
    currentUser = await api("/api/session");
  } catch (error) {
    if (error.status !== 401) showApiError(`Не удалось проверить сессию: ${error.message}`);
  }
  await refreshData();
  audioPlayer.volume = Number(document.querySelector("#volume-input").value) / 100;
  window.setInterval(() => refreshData({ silent: true }), 10000);
}

render();
initialize();
