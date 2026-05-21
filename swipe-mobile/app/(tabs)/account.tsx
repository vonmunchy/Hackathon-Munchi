import { useState, useEffect } from "react";
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ScrollView,
  ActivityIndicator,
  Alert,
  Linking,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import { useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

type SellerType = "marketplace" | "crypto" | "both";

export default function AccountScreen() {
  const [sessionToken, setSessionToken] = useState<string | null>(null);
  const [storeId, setStoreId] = useState<string | null>(null);
  const [showOnboarding, setShowOnboarding] = useState(false);

  if (showOnboarding && sessionToken) {
    return (
      <OnboardingFlow
        sessionToken={sessionToken}
        onComplete={() => setShowOnboarding(false)}
        onBack={() => {
          setShowOnboarding(false);
          setSessionToken(null);
          setStoreId(null);
        }}
      />
    );
  }

  if (sessionToken) {
    return (
      <SellerDashboard
        sessionToken={sessionToken}
        onLogout={() => {
          setSessionToken(null);
          setStoreId(null);
        }}
      />
    );
  }

  return (
    <LoginForm
      onLogin={(token: string, id: string, needsOnboarding: boolean) => {
        setSessionToken(token);
        setStoreId(id);
        if (needsOnboarding) setShowOnboarding(true);
      }}
    />
  );
}

/* ─── Login Form ─── */
function LoginForm({
  onLogin,
}: {
  onLogin: (token: string, storeId: string, needsOnboarding: boolean) => void;
}) {
  const [slug, setSlug] = useState("");
  const [pin, setPin] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const verifyPin = useQuery(
    api.stores.verifyPin,
    slug.length > 0 && pin.length === 4 ? { slug: slug.toLowerCase(), pin } : "skip"
  );
  const createSession = useMutation(api.sessions.create);

  const handleLogin = async () => {
    if (!slug.trim()) { setError("Enter your store slug"); return; }
    if (pin.length !== 4) { setError("PIN must be 4 digits"); return; }
    setLoading(true);
    setError("");

    try {
      if (!verifyPin) { setError("Verifying..."); setLoading(false); return; }
      if (!verifyPin.found) { setError("Store not found"); setLoading(false); return; }
      if (!verifyPin.pinMatch) { setError("Incorrect PIN"); setLoading(false); return; }
      if (!("storeId" in verifyPin) || !verifyPin.storeId) { setError("Login failed"); setLoading(false); return; }

      const token = generateToken();
      await createSession({ token, storeId: verifyPin.storeId });
      onLogin(token, verifyPin.storeId as string, false);
    } catch (err: any) {
      setError(err.message || "Login failed");
    } finally {
      setLoading(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      <ScrollView
        contentContainerStyle={{ flexGrow: 1, justifyContent: "center" }}
        className="px-6"
        keyboardShouldPersistTaps="handled"
      >
        {/* Branding */}
        <View className="items-center mb-10">
          <View className="h-16 w-16 rounded-2xl bg-amethyst-600 items-center justify-center mb-4">
            <Text className="text-white text-2xl font-bold">S</Text>
          </View>
          <Text className="text-2xl font-bold text-slate-900">Seller Portal</Text>
          <Text className="text-sm text-slate-500 mt-1">Sign in to manage your store</Text>
        </View>

        {/* Meta Business Login */}
        <View className="bg-white rounded-2xl border border-slate-100 p-6 mb-6 shadow-sm">
          <TouchableOpacity
            className="bg-[#0082FB] rounded-xl py-4 flex-row items-center justify-center"
            activeOpacity={0.8}
            onPress={() => Alert.alert(
              "Meta Business Login",
              "Meta OAuth requires the web app. Use the PIN login below for the mobile demo, or visit the web app to connect your Meta account."
            )}
          >
            <Ionicons name="logo-facebook" size={20} color="white" />
            <Text className="text-white font-semibold text-base ml-2">
              Continue with Meta Business
            </Text>
          </TouchableOpacity>
          <Text className="text-center text-xs text-slate-400 mt-3">
            Connect your Facebook Page and Instagram
          </Text>
        </View>

        {/* Divider */}
        <View className="flex-row items-center mb-6">
          <View className="flex-1 h-px bg-slate-200" />
          <Text className="mx-4 text-xs text-slate-400 font-medium">OR</Text>
          <View className="flex-1 h-px bg-slate-200" />
        </View>

        {/* PIN Login */}
        <View className="bg-white rounded-2xl border border-slate-100 p-6 shadow-sm">
          <View className="mb-4">
            <Text className="text-sm font-medium text-slate-700 mb-1.5">Store Slug</Text>
            <TextInput
              className="border border-slate-200 rounded-xl px-4 py-3.5 text-base text-slate-900 bg-slate-50"
              placeholder="e.g. my-store"
              placeholderTextColor="#94a3b8"
              value={slug}
              onChangeText={(t) => { setSlug(t.toLowerCase().replace(/[^a-z0-9-]/g, "")); setError(""); }}
              autoCapitalize="none"
              autoCorrect={false}
            />
          </View>

          <View className="mb-4">
            <Text className="text-sm font-medium text-slate-700 mb-1.5">PIN</Text>
            <TextInput
              className="border border-slate-200 rounded-xl px-4 py-3.5 text-base text-slate-900 bg-slate-50 tracking-[8px]"
              placeholder="••••"
              placeholderTextColor="#94a3b8"
              value={pin}
              onChangeText={(t) => { setPin(t.replace(/[^0-9]/g, "").slice(0, 4)); setError(""); }}
              keyboardType="number-pad"
              maxLength={4}
              secureTextEntry
            />
          </View>

          {error ? (
            <View className="bg-red-50 rounded-lg px-3 py-2 mb-3">
              <Text className="text-red-600 text-sm">{error}</Text>
            </View>
          ) : null}

          <TouchableOpacity
            className={`rounded-xl py-4 items-center ${loading ? "bg-amethyst-400" : "bg-amethyst-600"}`}
            onPress={handleLogin}
            disabled={loading}
            activeOpacity={0.8}
          >
            {loading ? (
              <ActivityIndicator color="white" />
            ) : (
              <Text className="text-white font-semibold text-base">Sign In</Text>
            )}
          </TouchableOpacity>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

/* ─── Onboarding Flow ─── */
function OnboardingFlow({
  sessionToken,
  onComplete,
  onBack,
}: {
  sessionToken: string;
  onComplete: () => void;
  onBack: () => void;
}) {
  const [step, setStep] = useState(0);
  const [sellerType, setSellerType] = useState<SellerType | null>(null);
  const [phone, setPhone] = useState("");
  const [walletAddress, setWalletAddress] = useState("");
  const [clientId, setClientId] = useState("");
  const [clientSecret, setClientSecret] = useState("");
  const [importLoading, setImportLoading] = useState(false);
  const [imported, setImported] = useState(false);

  const session = useQuery(api.sessions.validate, { token: sessionToken });
  const setSellerTypeMutation = useMutation(api.stores.setSellerType);
  const setContactPhone = useMutation(api.stores.setContactPhone);
  const setSwipeCredentials = useMutation(api.stores.setSwipeCredentials);
  const setCryptoWallet = useMutation(api.stores.setCryptoWallet);
  const importTestProducts = useMutation(api.products.importTestProducts);
  const completeOnboarding = useMutation(api.stores.completeOnboarding);

  const storeName = session?.storeName ?? "Your Store";
  const isWalletValid = walletAddress.startsWith("T") && walletAddress.length === 34 && /^[A-Za-z0-9]+$/.test(walletAddress);

  const handleSelectPath = async (type: SellerType) => {
    setSellerType(type);
    await setSellerTypeMutation({ sessionToken, sellerType: type });
    setStep(1);
  };

  const handlePhoneNext = async () => {
    if (phone.length < 7) return;
    await setContactPhone({ sessionToken, contactPhone: `+960${phone}` });
    setStep(2);
  };

  const handleSwipeNext = async () => {
    if (!clientId || !clientSecret) return;
    await setSwipeCredentials({ sessionToken, swipeClientId: clientId, swipeClientSecret: clientSecret });
    setStep(3);
  };

  const handleImport = async () => {
    setImportLoading(true);
    try { await importTestProducts({ sessionToken }); setImported(true); } catch {}
    setImportLoading(false);
  };

  const handleProductsNext = () => {
    if (sellerType === "both") setStep(4);
    else handleFinish();
  };

  const handleWalletNext = async () => {
    if (!isWalletValid) return;
    await setCryptoWallet({ sessionToken, cryptoWalletAddress: walletAddress });
    handleFinish();
  };

  const handleFinish = async () => {
    await completeOnboarding({ sessionToken });
    onComplete();
  };

  const isCryptoWalletStep = (sellerType === "crypto" && step === 2) || (sellerType === "both" && step === 4);
  const isSwipeStep = sellerType !== "crypto" && step === 2;
  const isProductsStep = sellerType !== "crypto" && step === 3;

  const totalSteps = sellerType === "marketplace" ? 4 : sellerType === "crypto" ? 3 : 5;

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      <ScrollView className="flex-1 px-5" contentContainerStyle={{ paddingVertical: 24 }} keyboardShouldPersistTaps="handled">
        {/* Header */}
        <Text className="text-2xl font-bold text-slate-900 text-center">Set Up {storeName}</Text>
        {step > 0 && (
          <View className="mt-3">
            <Text className="text-sm text-slate-500 text-center mb-2">Step {step} of {totalSteps - 1}</Text>
            <View className="flex-row gap-1.5">
              {Array.from({ length: totalSteps - 1 }, (_, i) => (
                <View key={i} className={`flex-1 h-1.5 rounded-full ${i < step ? "bg-amethyst-500" : "bg-slate-200"}`} />
              ))}
            </View>
          </View>
        )}

        {/* Step 0: Path Selection */}
        {step === 0 && (
          <View className="mt-8 gap-4">
            <Text className="text-lg font-semibold text-slate-800 text-center">What would you like to do?</Text>
            <Text className="text-sm text-slate-500 text-center mb-2">Choose how you want to use your store</Text>

            <PathCard
              icon="bag-handle-outline"
              title="Sell Products"
              description="List and sell physical products via your storefront"
              onPress={() => handleSelectPath("marketplace")}
            />
            <PathCard
              icon="lock-closed-outline"
              title="Sell Crypto"
              description="Sell USDT securely with instant Swipe payments"
              onPress={() => handleSelectPath("crypto")}
            />
            <PathCard
              icon="layers-outline"
              title="Sell Both"
              description="Access both marketplace and crypto exchange features"
              onPress={() => handleSelectPath("both")}
            />
          </View>
        )}

        {/* Step 1: Phone */}
        {step === 1 && (
          <StepCard title="Contact Number" subtitle="Your phone for customers and WhatsApp">
            <Text className="text-sm font-medium text-slate-700 mb-1.5">Phone</Text>
            <View className="flex-row">
              <View className="bg-slate-100 border border-slate-200 border-r-0 rounded-l-xl px-3 justify-center">
                <Text className="text-sm text-slate-500">+960</Text>
              </View>
              <TextInput
                className="flex-1 border border-slate-200 rounded-r-xl px-4 py-3 text-base text-slate-900 bg-white"
                placeholder="7XXXXXX"
                placeholderTextColor="#94a3b8"
                value={phone}
                onChangeText={(t) => setPhone(t.replace(/\D/g, "").slice(0, 7))}
                keyboardType="number-pad"
                maxLength={7}
              />
            </View>
            <StepButtons onBack={() => setStep(0)} onNext={handlePhoneNext} nextDisabled={phone.length < 7} />
          </StepCard>
        )}

        {/* Step 2: Swipe Credentials */}
        {isSwipeStep && (
          <StepCard title="Swipe API Credentials" subtitle="Enter your Swipe merchant credentials">
            <Text className="text-sm font-medium text-slate-700 mb-1.5">Client ID</Text>
            <TextInput
              className="border border-slate-200 rounded-xl px-4 py-3 text-sm text-slate-900 bg-white mb-3"
              placeholder="cli_..."
              placeholderTextColor="#94a3b8"
              value={clientId}
              onChangeText={setClientId}
              autoCapitalize="none"
            />
            <Text className="text-sm font-medium text-slate-700 mb-1.5">Client Secret</Text>
            <TextInput
              className="border border-slate-200 rounded-xl px-4 py-3 text-sm text-slate-900 bg-white"
              placeholder="sec_..."
              placeholderTextColor="#94a3b8"
              value={clientSecret}
              onChangeText={setClientSecret}
              secureTextEntry
            />
            <StepButtons onBack={() => setStep(1)} onNext={handleSwipeNext} nextDisabled={!clientId || !clientSecret} />
          </StepCard>
        )}

        {/* Step 3: Import Products */}
        {isProductsStep && (
          <StepCard title="Import Test Products" subtitle="Get started with 5 sample products">
            {[
              { name: "Black Abaya", price: "650" },
              { name: "Eid Gift Box", price: "450" },
              { name: "iPhone Case", price: "120" },
              { name: "Coral Bracelet", price: "85" },
              { name: "Premium Dates Box", price: "280" },
            ].map((p) => (
              <View key={p.name} className="flex-row items-center bg-slate-50 rounded-lg px-3 py-2.5 mb-1.5 border border-slate-100">
                <Text className="flex-1 text-sm font-medium text-slate-800">{p.name}</Text>
                <Text className="text-sm text-amethyst-600 font-mono">MVR {p.price}</Text>
              </View>
            ))}
            {!imported ? (
              <TouchableOpacity
                className={`rounded-xl py-3.5 items-center mt-3 ${importLoading ? "bg-slate-300" : "bg-ruby-500"}`}
                onPress={handleImport}
                disabled={importLoading}
              >
                <Text className="text-white font-semibold">{importLoading ? "Importing..." : "Import All Test Products"}</Text>
              </TouchableOpacity>
            ) : (
              <View className="bg-green-50 rounded-lg px-3 py-2.5 mt-3">
                <Text className="text-green-700 text-sm font-medium text-center">5 products imported!</Text>
              </View>
            )}
            <StepButtons
              onBack={() => setStep(2)}
              onNext={handleProductsNext}
              nextLabel={sellerType === "both" ? "Next" : imported ? "Go to Dashboard" : "Skip & Finish"}
            />
          </StepCard>
        )}

        {/* Crypto Wallet Step */}
        {isCryptoWalletStep && (
          <StepCard title="USDT Wallet (TRC20)" subtitle="Where you'll receive withdrawals">
            <Text className="text-sm font-medium text-slate-700 mb-1.5">Wallet Address</Text>
            <TextInput
              className={`border rounded-xl px-4 py-3 text-sm text-slate-900 bg-white ${
                walletAddress.length > 0
                  ? isWalletValid ? "border-green-400" : "border-red-400"
                  : "border-slate-200"
              }`}
              placeholder="T..."
              placeholderTextColor="#94a3b8"
              value={walletAddress}
              onChangeText={(t) => setWalletAddress(t.replace(/[^A-Za-z0-9]/g, "").slice(0, 34))}
              autoCapitalize="characters"
              maxLength={34}
            />
            {walletAddress.length > 0 && !isWalletValid && (
              <Text className="text-xs text-red-500 mt-1">
                {!walletAddress.startsWith("T") ? "Must start with T" : `${walletAddress.length}/34 characters`}
              </Text>
            )}
            <Text className="text-xs text-slate-400 mt-1">Example: TJfKxBkqz1x5Qp7ZxkLkg1F5DwLQGpV8Br</Text>
            <StepButtons
              onBack={() => setStep(sellerType === "both" ? 3 : 1)}
              onNext={handleWalletNext}
              nextDisabled={!isWalletValid}
              nextLabel="Continue"
            />
          </StepCard>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

/* ─── Seller Dashboard ─── */
function SellerDashboard({
  sessionToken,
  onLogout,
}: {
  sessionToken: string;
  onLogout: () => void;
}) {
  const router = useRouter();
  const stats = useQuery(api.orders.getDashboardStats, { sessionToken });
  const orders = useQuery(api.orders.listByStore, { sessionToken });
  const session = useQuery(api.sessions.validate, { token: sessionToken });
  const removeSession = useMutation(api.sessions.remove);

  const handleLogout = async () => {
    try { await removeSession({ token: sessionToken }); } catch {}
    onLogout();
  };

  if (stats === undefined || session === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center" edges={["top"]}>
        <ActivityIndicator size="large" color="#9333ea" />
        <Text className="text-slate-500 mt-3">Loading dashboard...</Text>
      </SafeAreaView>
    );
  }

  if (!session) { onLogout(); return null; }

  const recentOrders = orders?.slice(0, 5) ?? [];

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      <ScrollView className="flex-1" contentContainerStyle={{ paddingBottom: 40 }}>
        {/* Header */}
        <View className="px-5 pt-3 pb-5 flex-row items-center justify-between">
          <View className="flex-1">
            <Text className="text-2xl font-bold text-slate-900">{session.storeName}</Text>
            <Text className="text-sm text-slate-500 mt-0.5">Seller Dashboard</Text>
          </View>
          <TouchableOpacity
            className="bg-slate-100 rounded-full p-2.5"
            onPress={() =>
              Alert.alert("Logout", "Are you sure?", [
                { text: "Cancel", style: "cancel" },
                { text: "Logout", style: "destructive", onPress: handleLogout },
              ])
            }
          >
            <Ionicons name="log-out-outline" size={20} color="#64748b" />
          </TouchableOpacity>
        </View>

        {/* Stats */}
        <View className="px-5 flex-row flex-wrap gap-3">
          <StatCard label="Total Sales" value={`${(stats?.totalSalesAllTime ?? 0).toLocaleString()} MVR`} icon="wallet-outline" color="#7e22ce" bg="#faf5ff" />
          <StatCard label="Orders Paid" value={String(stats?.paidCount ?? 0)} icon="checkmark-circle-outline" color="#15803d" bg="#f0fdf4" />
          <StatCard label="Awaiting Ship" value={String(stats?.awaitingShipmentCount ?? 0)} icon="time-outline" color="#b45309" bg="#fffbeb" />
          <StatCard label="Low Stock" value={String(stats?.lowStockCount ?? 0)} icon="alert-circle-outline" color="#b91c1c" bg="#fef2f2" />
        </View>

        {/* Quick Actions */}
        <View className="px-5 mt-6 flex-row gap-3">
          <TouchableOpacity
            className="flex-1 bg-amethyst-600 rounded-2xl py-4 items-center flex-row justify-center gap-2"
            activeOpacity={0.8}
            onPress={() => router.push({ pathname: "/seller/products", params: { sessionToken, storeId: session.storeId } })}
          >
            <Ionicons name="cube-outline" size={18} color="white" />
            <Text className="text-white font-semibold text-sm">Products</Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="flex-1 bg-slate-800 rounded-2xl py-4 items-center flex-row justify-center gap-2"
            activeOpacity={0.8}
            onPress={() => router.push({ pathname: "/seller/orders", params: { sessionToken } })}
          >
            <Ionicons name="receipt-outline" size={18} color="white" />
            <Text className="text-white font-semibold text-sm">Orders</Text>
          </TouchableOpacity>
        </View>

        {/* Recent Orders */}
        <View className="px-5 mt-6">
          <Text className="text-lg font-bold text-slate-900 mb-3">Recent Orders</Text>
          {recentOrders.length === 0 ? (
            <View className="bg-white rounded-2xl p-8 items-center border border-slate-100">
              <Ionicons name="receipt-outline" size={32} color="#cbd5e1" />
              <Text className="text-slate-400 mt-2 text-sm">No orders yet</Text>
            </View>
          ) : (
            recentOrders.map((order) => (
              <View key={order._id} className="bg-white rounded-xl p-4 mb-2 border border-slate-100">
                <View className="flex-row items-center justify-between">
                  <View className="flex-1">
                    <Text className="text-sm font-semibold text-slate-900">{order.customerName}</Text>
                    <Text className="text-xs text-slate-400 mt-0.5">{new Date(order.createdAt).toLocaleDateString()}</Text>
                  </View>
                  <View className="items-end">
                    <Text className="text-sm font-bold text-slate-900">{order.totalAmount.toFixed(2)} {order.currency}</Text>
                    <StatusBadge status={order.status} />
                  </View>
                </View>
              </View>
            ))
          )}
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}

/* ─── Reusable Components ─── */

function PathCard({ icon, title, description, onPress }: {
  icon: keyof typeof Ionicons.glyphMap;
  title: string;
  description: string;
  onPress: () => void;
}) {
  return (
    <TouchableOpacity
      className="bg-white rounded-2xl border-2 border-slate-200 p-5 flex-row items-start gap-4"
      activeOpacity={0.7}
      onPress={onPress}
    >
      <View className="h-12 w-12 rounded-xl bg-amethyst-50 items-center justify-center">
        <Ionicons name={icon} size={24} color="#9333ea" />
      </View>
      <View className="flex-1">
        <Text className="text-base font-semibold text-slate-900">{title}</Text>
        <Text className="text-sm text-slate-500 mt-0.5">{description}</Text>
      </View>
      <Ionicons name="chevron-forward" size={20} color="#cbd5e1" />
    </TouchableOpacity>
  );
}

function StepCard({ title, subtitle, children }: {
  title: string;
  subtitle: string;
  children: React.ReactNode;
}) {
  return (
    <View className="mt-8 bg-white rounded-2xl border border-slate-200 p-5">
      <Text className="text-lg font-semibold text-slate-800">{title}</Text>
      <Text className="text-sm text-slate-500 mt-1 mb-5">{subtitle}</Text>
      {children}
    </View>
  );
}

function StepButtons({ onBack, onNext, nextDisabled, nextLabel }: {
  onBack: () => void;
  onNext: () => void;
  nextDisabled?: boolean;
  nextLabel?: string;
}) {
  return (
    <View className="flex-row gap-3 mt-5">
      <TouchableOpacity className="flex-1 border-2 border-slate-200 rounded-xl py-3.5 items-center" onPress={onBack}>
        <Text className="text-slate-600 font-semibold">Back</Text>
      </TouchableOpacity>
      <TouchableOpacity
        className={`flex-1 rounded-xl py-3.5 items-center ${nextDisabled ? "bg-slate-200" : "bg-amethyst-500"}`}
        onPress={onNext}
        disabled={nextDisabled}
      >
        <Text className={`font-semibold ${nextDisabled ? "text-slate-400" : "text-white"}`}>{nextLabel ?? "Next"}</Text>
      </TouchableOpacity>
    </View>
  );
}

function StatCard({ label, value, icon, color, bg }: {
  label: string; value: string; icon: keyof typeof Ionicons.glyphMap; color: string; bg: string;
}) {
  return (
    <View className="rounded-2xl p-4 border border-slate-100" style={{ width: "47%", backgroundColor: bg }}>
      <Ionicons name={icon} size={20} color={color} />
      <Text className="text-xl font-bold mt-2" style={{ color }}>{value}</Text>
      <Text className="text-xs text-slate-500 mt-1">{label}</Text>
    </View>
  );
}

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { bg: string; text: string }> = {
    pending: { bg: "#fef9c3", text: "#a16207" },
    paid: { bg: "#dcfce7", text: "#15803d" },
    shipped: { bg: "#dbeafe", text: "#1d4ed8" },
    delivered: { bg: "#f1f5f9", text: "#475569" },
    cancelled: { bg: "#fee2e2", text: "#b91c1c" },
  };
  const c = config[status] ?? config.pending;
  return (
    <View className="px-2 py-0.5 rounded-full mt-1" style={{ backgroundColor: c.bg }}>
      <Text className="text-xs font-medium capitalize" style={{ color: c.text }}>{status}</Text>
    </View>
  );
}

function generateToken(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let token = "";
  for (let i = 0; i < 32; i++) token += chars.charAt(Math.floor(Math.random() * chars.length));
  return token;
}
