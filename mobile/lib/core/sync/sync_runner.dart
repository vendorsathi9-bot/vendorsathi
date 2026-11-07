import 'dart:convert';
import '../api/api_client.dart';
import '../database/app_database.dart';

class SyncRunner {
  final ApiClient apiClient;

  SyncRunner({required this.apiClient});

  Future<int> processQueue() async {
    if (apiClient.token == null) return 0;

    final pending = await AppDatabase.getPendingOperations();
    int syncedCount = 0;

    for (final op in pending) {
      final clientOpId = op['client_operation_id'] as String;
      final deviceId = op['device_id'] as String;
      final opType = op['operation_type'] as String;
      final payloadJson = op['payload_json'] as String;

      try {
        final payloadMap = jsonDecode(payloadJson);
        final response = await apiClient.post('/sync/operations', body: {
          'client_operation_id': clientOpId,
          'device_id': deviceId,
          'operation_type': opType,
          'payload_version': 1,
          'payload': payloadMap,
        });

        await AppDatabase.markOperationStatus(
          clientOpId,
          response['status'] ?? 'ACCEPTED',
          resultJson: jsonEncode(response['result'] ?? {}),
        );
        syncedCount++;
      } catch (e) {
        // Leave as PENDING or mark error for retry
        break;
      }
    }

    return syncedCount;
  }
}
