import { useState } from "react";
import { StatusBar } from "expo-status-bar";
import { StyleSheet, Text, View, Pressable, Alert, TextInput } from "react-native";
import { CameraView, useCameraPermissions } from "expo-camera";

const API_BASE = "http://10.0.2.2:4000";
// 10.0.2.2 = host loopback from Android emulator. On a real device or iOS use your machine LAN IP.

type Screen = "login" | "home" | "scanner" | "confirm" | "success";

export default function App() {
  const [screen, setScreen] = useState<Screen>("login");
  const [customer, setCustomer] = useState("Ali Hassan");
  const [permission, requestPermission] = useCameraPermissions();
  const [scanned, setScanned] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function openScanner() {
    if (!permission?.granted) {
      const r = await requestPermission();
      if (!r.granted) {
        Alert.alert("Camera needed to scan ATM QR codes.");
        return;
      }
    }
    setScanned(null);
    setScreen("scanner");
  }

  async function approve() {
    if (!scanned) return;
    setBusy(true);
    try {
      const r = await fetch(`${API_BASE}/api/session/approve`, {
        method: "POST",
        headers: { "content-type": "application/json" },
        body: JSON.stringify({
          qrToken: scanned,
          deviceId: "demo-device-001",
          customerRef: customer,
        }),
      });
      const j = await r.json();
      if (!r.ok) {
        Alert.alert("Could not authorise", j.error ?? "unknown");
        setScreen("home");
        return;
      }
      setScreen("success");
      setTimeout(() => setScreen("home"), 3000);
    } catch (e: any) {
      Alert.alert("Network error", String(e?.message ?? e));
      setScreen("home");
    } finally {
      setBusy(false);
    }
  }

  return (
    <View style={styles.root}>
      <StatusBar style="light" />

      {screen === "login" && (
        <View style={styles.card}>
          <Text style={styles.brand}>DemoBank</Text>
          <Text style={styles.label}>Customer name (demo)</Text>
          <TextInput style={styles.input} value={customer} onChangeText={setCustomer} />
          <Pressable style={styles.btnPrimary} onPress={() => setScreen("home")}>
            <Text style={styles.btnText}>Log in with Face ID (mocked)</Text>
          </Pressable>
        </View>
      )}

      {screen === "home" && (
        <View style={styles.card}>
          <Text style={styles.brand}>DemoBank</Text>
          <Text style={styles.hello}>Hi {customer}</Text>
          <Text style={styles.balance}>£ 2,431.22</Text>

          <Pressable style={styles.tile} onPress={openScanner}>
            <Text style={styles.tileTitle}>Cardless ATM Access</Text>
            <Text style={styles.tileSub}>Scan the QR on an ATM to withdraw without your card</Text>
          </Pressable>
          <Pressable style={[styles.tile, styles.disabled]}>
            <Text style={styles.tileTitle}>Transfer</Text>
          </Pressable>
          <Pressable style={[styles.tile, styles.disabled]}>
            <Text style={styles.tileTitle}>Statements</Text>
          </Pressable>
        </View>
      )}

      {screen === "scanner" && (
        <View style={{ flex: 1 }}>
          <CameraView
            style={{ flex: 1 }}
            facing="back"
            barcodeScannerSettings={{ barcodeTypes: ["qr"] }}
            onBarcodeScanned={(r) => {
              if (scanned) return;
              setScanned(r.data);
              setScreen("confirm");
            }}
          />
          <Pressable style={styles.cancelBar} onPress={() => setScreen("home")}>
            <Text style={styles.btnText}>Cancel</Text>
          </Pressable>
        </View>
      )}

      {screen === "confirm" && (
        <View style={styles.card}>
          <Text style={styles.brand}>Approve ATM access?</Text>
          <Text style={styles.tileSub}>
            You're about to unlock an ATM session for cash withdrawal. This does not move any money — the ATM will
            ask you for an amount next.
          </Text>
          <Pressable style={styles.btnPrimary} onPress={approve} disabled={busy}>
            <Text style={styles.btnText}>{busy ? "Authorising…" : "Approve with biometrics (mocked)"}</Text>
          </Pressable>
          <Pressable style={styles.btnGhost} onPress={() => setScreen("home")}>
            <Text style={styles.btnTextDark}>Cancel</Text>
          </Pressable>
        </View>
      )}

      {screen === "success" && (
        <View style={styles.card}>
          <Text style={styles.brand}>ATM unlocked</Text>
          <Text style={styles.tileSub}>You can now use the ATM. This screen will close automatically.</Text>
        </View>
      )}
    </View>
  );
}

const C = {
  sky: "#447794",
  steel: "#2D5B75",
  deep: "#123249",
  night: "#061222",
  ink: "#F5F8FB",
  muted: "#8FB1C4",
  line: "rgba(68, 119, 148, 0.25)",
  accent: "#C9A961",
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: C.night },
  card: { flex: 1, padding: 28, paddingTop: 90, gap: 16, backgroundColor: C.night },
  brand: { color: C.sky, fontSize: 13, fontWeight: "700", letterSpacing: 4, textTransform: "uppercase" },
  hello: { color: C.ink, fontSize: 26, marginTop: 12, fontWeight: "500" },
  balance: { color: C.ink, fontSize: 42, fontWeight: "700", marginBottom: 20, letterSpacing: -1 },
  label: { color: C.muted, marginTop: 14, fontSize: 13, letterSpacing: 1, textTransform: "uppercase" },
  input: {
    backgroundColor: "rgba(68, 119, 148, 0.08)",
    color: C.ink,
    padding: 16,
    borderRadius: 10,
    borderWidth: 1,
    borderColor: C.line,
    fontSize: 16,
  },
  tile: {
    backgroundColor: "rgba(68, 119, 148, 0.08)",
    padding: 20,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: C.line,
    gap: 6,
  },
  tileTitle: { color: C.ink, fontSize: 17, fontWeight: "600" },
  tileSub: { color: C.muted, lineHeight: 21, fontSize: 14 },
  disabled: { opacity: 0.35 },
  btnPrimary: {
    backgroundColor: C.sky,
    padding: 18,
    borderRadius: 12,
    alignItems: "center",
    marginTop: 16,
  },
  btnGhost: { padding: 14, alignItems: "center" },
  btnText: { color: C.night, fontSize: 16, fontWeight: "700", letterSpacing: 0.3 },
  btnTextDark: { color: C.muted, fontSize: 15 },
  cancelBar: { backgroundColor: C.night, padding: 20, alignItems: "center" },
});
