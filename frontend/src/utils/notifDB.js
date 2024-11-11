const dbName = "NotifDB";
const dbVersion = 1;

let db;

const openDB = () => {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(dbName, dbVersion);

    request.onerror = (event) => {
      console.error("Database error: ", event.target.errorCode);
      reject(event.target.errorCode);
    };

    request.onsuccess = (event) => {
      console.log("Database opened successfully");
      db = event.target.result;
      resolve(db);
    };

    request.onupgradeneeded = (event) => {
      db = event.target.result;
      if (!db.objectStoreNames.contains("chats")) {
        const chatStore = db.createObjectStore("chats", { keyPath: "chat_id" });
        chatStore.createIndex("chatList", "chatList", { unique: false });
        chatStore.createIndex("messageCounter", "messageCounter", { unique: false });
      }
      if (!db.objectStoreNames.contains("notifications")) {
        const notificationStore = db.createObjectStore("notifications", { keyPath: "id" });
        notificationStore.createIndex("notificationList", "notificationList", { unique: false });
        notificationStore.createIndex("notificationCounter", "notificationCounter", { unique: false });
      }
      console.log("Database setup complete");
    };
  });
};

const getDB = async () => {
  if (!db) {
    db = await openDB();
  }
  return db;
};

const addData = async (storeName, data) => {
  const db = await getDB();
  const transaction = db.transaction([storeName], "readwrite");
  const store = transaction.objectStore(storeName);

  // Convert arrays and objects to JSON strings before storing
  const serializedData = JSON.parse(JSON.stringify(data));

  const request = store.put(serializedData);

  request.onsuccess = () => {
    console.log(`${storeName} data added to store`);
  };

  request.onerror = (event) => {
    console.error(`Error adding ${storeName} data to store:`, event.target.errorCode);
  };
};

const getData = async (storeName, id) => {
  const db = await getDB();
  return new Promise((resolve, reject) => {
    const transaction = db.transaction([storeName]);
    const store = transaction.objectStore(storeName);
    const request = store.get(id);

    request.onsuccess = (event) => {
      // Parse JSON strings back into arrays and objects
      const result = event.target.result ? JSON.parse(JSON.stringify(event.target.result)) : null;
      resolve(result);
    };

    request.onerror = (event) => {
      reject(event.target.errorCode);
    };
  });
};

export { openDB, addData, getData };
