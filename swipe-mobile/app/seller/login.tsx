import { useState } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  Alert,
  TextInput,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter } from "expo-router";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import { Ionicons } from "@expo/vector-icons";

export default function SellerLoginScreen() {
  const router = useRouter();
  const [showDemoLogin, setShowDemoLogin] = useState(false);
  const [slug, setSlug] = useState("");
  const [pin, setPin] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const storeResult = useQuery(
    api.stores.verifyPin,
    slug.length > 0 && pin.length === 4 ? { slug, pin } : "skip"
  );

  const createSession = useMutation(api.sessions.create);

  const handleMetaLogin = () => {
    Alert.alert(
      "Meta Business Login",
      "OAuth login requires a browser. Use the web app at swipe.mv/seller for Meta login.\n\nFor the hackathon demo, use the demo login instead.",
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Demo Login",
          onPress: () => setShowDemoLogin(true),
        },
      ]
    );
  };

  const handleDemoLogin = async () => {
    if (!slug.trim() || pin.length !== 4) {
      Alert.alert("Error", "Enter your store slug and 4-digit PIN");
      return;
    }

    setIsLoading(true);
    try {
      // storeResult is reactive — but we need to query directly for the action
      // Use the current storeResult if available
      if (!storeResult) {
        Alert.alert("Error", "Verifying store... please try again.");
        setIsLoading(false);
        return;
      }

      if (!storeResult.found) {
        Alert.alert("Error", "Store not found. Check your slug.");
        setIsLoading(false);
        return;
      }

      if (!storeResult.pinMatch) {
        Alert.alert("Error", "Incorrect PIN.");
        setIsLoading(false);
        return;
      }

      // Create session
      const token = `mobile_${Date.now()}_${Math.random().toString(36).slice(2)}`;
      await createSession({
        token,
        storeId: storeResult.storeId!,
      });

      Alert.alert("Welcome!", "Logged in successfully");
      router.replace({
        pathname: "/seller/dashboard",
        params: { sessionToken: token },
      });
    } catch (error: any) {
      Alert.alert("Error", error.message || "Login failed");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <SafeAreaView className="flex-1 bg-slate-50">
      {/* Top Bar */}
      <View className="flex-row items-center px-4 py-3 border-b border-slate-100 bg-white">
        <TouchableOpacity
          onPress={() => router.back()}
          className="h-10 w-10 items-center justify-center rounded-xl"
          activeOpacity={0.7}
        >
          <Ionicons name="arrow-back" size={22} color="#334155" />
        </TouchableOpacity>
        <Text className="text-lg font-semibold text-slate-900 ml-2">
          Swipe Store
        </Text>
      </View>

      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        className="flex-1"
      >
        <ScrollView
          contentContainerStyle={{ flexGrow: 1, justifyContent: "center" }}
          keyboardShouldPersistTaps="handled"
        >
          {/* Centered Content */}
          <View className="items-center px-6">
            {/* Purple S Icon */}
            <View className="h-16 w-16 rounded-2xl bg-violet-600 items-center justify-center mb-6">
              <Text className="text-white text-2xl font-bold">S</Text>
            </View>

            {/* Heading */}
            <Text className="text-2xl font-bold text-slate-900 mb-1">
              Seller Portal
            </Text>
            <Text className="text-sm text-slate-500 mb-8">
              Sign in to manage your store
            </Text>

            {/* Card */}
            <View className="w-full bg-white rounded-xl border border-slate-200 p-5">
              {/* Meta Login Button */}
              <TouchableOpacity
                className="flex-row items-center justify-center py-3.5 rounded-xl"
                style={{ backgroundColor: "#0082FB" }}
                activeOpacity={0.8}
                onPress={handleMetaLogin}
              >
                <Ionicons name="logo-facebook" size={20} color="#ffffff" />
                <Text className="text-white font-semibold text-sm ml-2">
                  Continue with Meta Business
                </Text>
              </TouchableOpacity>

              <Text className="text-xs text-slate-400 text-center mt-3 leading-4">
                Connect your Facebook Page and Instagram to manage your store
              </Text>

              {/* Demo Login Form (shown after tapping Meta button) */}
              {showDemoLogin && (
                <View className="mt-5 pt-5 border-t border-slate-100">
                  <View className="flex-row items-center mb-4">
                    <View className="h-5 w-5 rounded-full bg-amber-100 items-center justify-center mr-2">
                      <Ionicons name="flash" size={12} color="#d97706" />
                    </View>
                    <Text className="text-sm font-medium text-slate-700">
                      Hackathon Demo Login
                    </Text>
                  </View>

                  {/* Slug Input */}
                  <Text className="text-xs font-medium text-slate-500 mb-1.5">
                    Store Slug
                  </Text>
                  <TextInput
                    className="border border-slate-200 rounded-xl px-4 py-3 text-base text-slate-900 mb-3"
                    placeholder="e.g. island-tees"
                    placeholderTextColor="#94a3b8"
                    value={slug}
                    onChangeText={setSlug}
                    autoCapitalize="none"
                    autoCorrect={false}
                  />

                  {/* PIN Input */}
                  <Text className="text-xs font-medium text-slate-500 mb-1.5">
                    4-Digit PIN
                  </Text>
                  <TextInput
                    className="border border-slate-200 rounded-xl px-4 py-3 text-base text-slate-900 mb-4"
                    placeholder="0000"
                    placeholderTextColor="#94a3b8"
                    value={pin}
                    onChangeText={(t) => setPin(t.replace(/[^0-9]/g, "").slice(0, 4))}
                    keyboardType="number-pad"
                    maxLength={4}
                    secureTextEntry
                  />

                  {/* Login Button */}
                  <TouchableOpacity
                    className={`rounded-xl py-3.5 items-center ${
                      slug.trim() && pin.length === 4 && !isLoading
                        ? "bg-violet-600"
                        : "bg-slate-200"
                    }`}
                    disabled={!slug.trim() || pin.length !== 4 || isLoading}
                    activeOpacity={0.8}
                    onPress={handleDemoLogin}
                  >
                    {isLoading ? (
                      <ActivityIndicator size="small" color="#ffffff" />
                    ) : (
                      <Text
                        className={`font-semibold text-sm ${
                          slug.trim() && pin.length === 4
                            ? "text-white"
                            : "text-slate-400"
                        }`}
                      >
                        Demo Login
                      </Text>
                    )}
                  </TouchableOpacity>
                </View>
              )}
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
