import { View, Text, ScrollView, TouchableOpacity } from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

export default function HomeScreen() {
  const router = useRouter();

  return (
    <SafeAreaView className="flex-1 bg-white" edges={["top"]}>
      <ScrollView
        className="flex-1"
        contentContainerStyle={{ paddingBottom: 48 }}
        showsVerticalScrollIndicator={false}
      >
        {/* Nav bar */}
        <View className="flex-row items-center px-5 pt-4 pb-3 border-b border-slate-100">
          <View className="w-8 h-8 rounded-full bg-[#9333ea] items-center justify-center">
            <Text className="text-white text-xs font-bold">S</Text>
          </View>
          <Text className="ml-2.5 text-lg font-bold text-slate-900">
            Swipe
          </Text>
        </View>

        {/* Hero section */}
        <View className="px-5 pt-10">
          {/* Powered by Swipe badge */}
          <View className="flex-row items-center self-start rounded-full bg-[#9333ea]/5 border border-[#9333ea]/10 px-3.5 py-1.5 mb-6">
            <View className="w-1.5 h-1.5 rounded-full bg-[#9333ea] mr-2" />
            <Text className="text-xs font-semibold text-[#7e22ce] tracking-wide">
              Powered by Swipe
            </Text>
          </View>

          {/* Hero heading */}
          <Text className="text-[28px] font-extrabold text-slate-900 leading-[1.1] tracking-tight">
            What wasn't possible before — is now possible with Swipe
          </Text>

          {/* Subtitle */}
          <Text className="mt-5 text-[15px] text-slate-500 leading-relaxed">
            Shop from local sellers or buy USDT — all with instant Swipe
            payments. No bank transfers. No scams. No waiting.
          </Text>

          {/* CTA buttons */}
          <View className="mt-8">
            <TouchableOpacity
              onPress={() => router.push("/(tabs)/marketplace")}
              className="flex-row items-center justify-center rounded-xl bg-[#9333ea] px-6 py-3.5 shadow-sm mb-3"
              activeOpacity={0.85}
            >
              <Text className="text-sm font-semibold text-white mr-2">
                Shop Marketplace
              </Text>
              <Ionicons name="arrow-forward" size={16} color="#ffffff" />
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => router.push("/(tabs)/exchange")}
              className="flex-row items-center justify-center rounded-xl bg-slate-800 px-6 py-3.5 shadow-sm"
              activeOpacity={0.85}
            >
              <Text className="text-sm font-semibold text-white mr-2">
                Buy Crypto
              </Text>
              <Ionicons name="arrow-forward" size={16} color="#ffffff" />
            </TouchableOpacity>
          </View>

          {/* Text links */}
          <View className="flex-row mt-5 gap-4">
            <TouchableOpacity
              onPress={() => router.push("/seller/login")}
              activeOpacity={0.7}
            >
              <Text className="text-sm font-medium text-slate-500 underline underline-offset-4 decoration-slate-300">
                Sell Products
              </Text>
            </TouchableOpacity>

            <TouchableOpacity
              onPress={() => router.push("/seller/login")}
              activeOpacity={0.7}
            >
              <Text className="text-sm font-medium text-slate-500 underline underline-offset-4 decoration-slate-300">
                Sell Crypto
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </ScrollView>
    </SafeAreaView>
  );
}
