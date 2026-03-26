// ---------------------------------------------------------------
// mongo-init.js
// Runs once on the very first start of the MongoDB container.
// ---------------------------------------------------------------

// Switch to (or create) the target database.
var db = db.getSiblingDB("person");

// Explicitly create the collection.  This is optional – the
// collection would also be created automatically on first insert,
// but using `createCollection` guarantees it exists with default
// options (no capped, no validator, …).
if (!db.getCollectionNames().includes("person")) {
    db.createCollection("person");
}

// (Optional) Create a simple index on the "id" field so that
// look‑ups by `_id` are fast.  Mongo already indexes `_id`, but if you
// ever decide to store an additional logical “id” field you can
// uncomment the line below.
//
// db.person.createIndex({ id: 1 }, { unique: true });
