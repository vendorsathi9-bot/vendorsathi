import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';

class AppDatabase {
  static Database? _database;

  static Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDB('vendorsaathi_local.db');
    return _database!;
  }

  static Future<Database> _initDB(String filePath) async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, filePath);

    return await openDatabase(
      path,
      version: 1,
      onCreate: _createDB,
    );
  }

  static Future<void> _createDB(Database db, int version) async {
    // Cached products
    await db.execute('''
      CREATE TABLE cache_products (
        id TEXT PRIMARY KEY,
        vendor_id TEXT NOT NULL,
        name TEXT NOT NULL,
        description TEXT,
        price_paise INTEGER NOT NULL,
        currency TEXT NOT NULL,
        active INTEGER NOT NULL,
        updated_at TEXT NOT NULL
      )
    ''');

    // Cached inventory
    await db.execute('''
      CREATE TABLE cache_inventory (
        product_id TEXT PRIMARY KEY,
        on_hand_qty INTEGER NOT NULL,
        reserved_qty INTEGER NOT NULL,
        available_qty INTEGER NOT NULL,
        updated_at TEXT NOT NULL
      )
    ''');

    // Offline operations outbox
    await db.execute('''
      CREATE TABLE outbox_operations (
        id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        device_id TEXT NOT NULL,
        client_operation_id TEXT NOT NULL UNIQUE,
        operation_type TEXT NOT NULL,
        payload_version INTEGER NOT NULL,
        status TEXT NOT NULL,
        payload_json TEXT NOT NULL,
        result_json TEXT,
        created_at TEXT NOT NULL,
        processed_at TEXT
      )
    ''');
  }

  // Outbox helpers
  static Future<void> queueOperation({
    required String id,
    required String userId,
    required String deviceId,
    required String clientOpId,
    required String opType,
    required String payloadJson,
  }) async {
    final db = await database;
    await db.insert('outbox_operations', {
      'id': id,
      'user_id': userId,
      'device_id': deviceId,
      'client_operation_id': clientOpId,
      'operation_type': opType,
      'payload_version': 1,
      'status': 'PENDING',
      'payload_json': payloadJson,
      'created_at': DateTime.now().toUtc().toIso8601String(),
    });
  }

  static Future<List<Map<String, dynamic>>> getPendingOperations() async {
    final db = await database;
    return await db.query(
      'outbox_operations',
      where: 'status = ?',
      whereArgs: ['PENDING'],
      orderBy: 'created_at ASC',
    );
  }

  static Future<void> markOperationStatus(String clientOpId, String status, {String? resultJson}) async {
    final db = await database;
    await db.update(
      'outbox_operations',
      {
        'status': status,
        'result_json': resultJson,
        'processed_at': DateTime.now().toUtc().toIso8601String(),
      },
      where: 'client_operation_id = ?',
      whereArgs: [clientOpId],
    );
  }
}
