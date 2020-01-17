const assert = require('assert').strict;

class RowClass {
  constructor(properties, options = {}) {
    const {
      etag,
      tableName,
      partitionKey,
      rowKey,
      db,
      context = {},
    } = options;

    assert(properties, 'properties is required');
    assert(tableName, 'tableName is required');
    assert(partitionKey, 'partitionKey is required');
    assert(rowKey, 'rowKey is required');
    assert(db, 'db is required');
    assert(typeof context === 'object' && context.constructor === Object, 'context should be an object');

    this.properties = properties;
    this.etag = etag;
    this.tableName = tableName;
    this.partitionKey = partitionKey;
    this.rowKey = rowKey;
    this.db = db;

    Object.entries(context).forEach(([key, value]) => {
      this[key] = value;
    });

    Object.entries(properties).forEach(([key, value]) => {
      this[key] = value;
    });
  }

  async remove(ignoreChanges, ignoreIfNotExists) {
    const [result] = await this.db.fns[`${this.tableName}_remove`](this.partitionKey, this.rowKey);

    if (result) {
      return true;
    }

    if (ignoreIfNotExists && !result) {
      return false;
    }

    if (!result) {
      const err = new Error('Resource not found');

      err.code = 'ResourceNotFound';
      err.statusCode = 404;

      throw err;
    }
  }

  // load the properties from the table once more, and return true if anything has changed.
  // Else, return false.
  async reload() {
    const result = await this.db.fns[`${this.tableName}_load`](this.partitionKey, this.rowKey);
    const etag = result[0].etag;

    return etag !== this.etag;
  }

  async modify(modifier) {
    await modifier.call(this.properties, this.properties);

    return this.db.fns[`${this.tableName}_modify`](this.partitionKey, this.rowKey, this.properties, 1);
  }
}

module.exports = RowClass;
