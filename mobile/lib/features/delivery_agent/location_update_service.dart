import 'dart:async';
import 'package:geolocator/geolocator.dart';
import '../../core/api/api_client.dart';

class LocationUpdateService {
  final ApiClient apiClient;
  Timer? _timer;
  String? _activeOrderId;
  bool _isBroadcasting = false;
  DateTime? _lastBroadcastTime;
  Position? _lastPosition;
  String? _lastError;

  Function(Position pos, DateTime time)? onLocationUpdated;
  Function(String error)? onError;

  LocationUpdateService({required this.apiClient});

  bool get isBroadcasting => _isBroadcasting;
  String? get activeOrderId => _activeOrderId;
  DateTime? get lastBroadcastTime => _lastBroadcastTime;
  Position? get lastPosition => _lastPosition;
  String? get lastError => _lastError;

  Future<bool> startBroadcasting(String orderId) async {
    _activeOrderId = orderId;
    _lastError = null;

    // Check location permissions
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      _lastError = 'Location services are disabled on device.';
      onError?.call(_lastError!);
      return false;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        _lastError = 'Location permissions denied by user.';
        onError?.call(_lastError!);
        return false;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      _lastError = 'Location permissions permanently denied.';
      onError?.call(_lastError!);
      return false;
    }

    _isBroadcasting = true;

    // Initial broadcast immediately
    await _sendUpdate();

    // Subsequent updates every 20 seconds
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 20), (timer) async {
      if (!_isBroadcasting || _activeOrderId == null) {
        timer.cancel();
        return;
      }
      await _sendUpdate();
    });

    return true;
  }

  Future<void> _sendUpdate() async {
    if (_activeOrderId == null) return;

    try {
      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 10),
      );

      _lastPosition = position;
      _lastBroadcastTime = DateTime.now();

      await apiClient.post(
        '/delivery/orders/$_activeOrderId/location',
        body: {
          'latitude': position.latitude,
          'longitude': position.longitude,
        },
      );

      onLocationUpdated?.call(position, _lastBroadcastTime!);
    } catch (e) {
      _lastError = 'Failed to broadcast GPS: $e';
      onError?.call(_lastError!);
    }
  }

  void stopBroadcasting() {
    _timer?.cancel();
    _timer = null;
    _isBroadcasting = false;
    _activeOrderId = null;
  }

  void dispose() {
    stopBroadcasting();
  }
}
