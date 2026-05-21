import { useState } from "react";
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  TextInput,
  Modal,
  KeyboardAvoidingView,
  Platform,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import type { Id } from "../../convex/_generated/dataModel";
import { Ionicons } from "@expo/vector-icons";

type Listing = {
  _id: Id<"exchangeListings">;
  storeName: string;
  usdtAmount: number;
  rate: number;
  availableBalance: number;
};

export default function ExchangeScreen() {
  const listings = useQuery(api.exchange.getActiveListings);
  const createReservation = useMutation(api.exchange.createReservation);
  const [selectedListing, setSelectedListing] = useState<Listing | null>(null);

  const processPurchase = async (listing: Listing, wallet: string) => {
    try {
      await createReservation({
        listingId: listing._id,
        amount: listing.availableBalance,
        buyerWallet: wallet,
      });
      Alert.alert(
        "Reservation Created!",
        `Reserved ${listing.availableBalance} USDT at ${listing.rate} MVR/USDT.\n\nTotal: ${(listing.availableBalance * listing.rate).toFixed(2)} MVR\n\nComplete payment via Swipe within 5 minutes.`
      );
    } catch (error: any) {
      Alert.alert("Error", error.message || "Failed to create reservation");
    }
  };

  const renderListing = ({ item }: { item: Listing }) => {
    const totalMvr = item.availableBalance * item.rate;
    return (
      <View className="mx-5 mb-3 rounded-xl bg-white p-5 border border-slate-200 shadow-sm">
        {/* Seller name */}
        <Text className="text-sm font-medium text-slate-500">
          {item.storeName}
        </Text>

        {/* USDT Amount */}
        <Text className="text-lg font-bold text-slate-900 mt-2">
          {item.availableBalance.toLocaleString()} USDT
        </Text>

        {/* Rate — inline with bold rate value */}
        <Text className="text-sm text-slate-600 mt-1">
          Rate:{" "}
          <Text className="font-bold text-slate-900">
            MVR {item.rate.toFixed(2)}
          </Text>
          <Text className="text-slate-400"> / USDT</Text>
        </Text>

        {/* Total — inline */}
        <Text className="text-sm text-slate-600 mt-0.5">
          Total: MVR {totalMvr.toFixed(2)}
        </Text>

        {/* Buy button */}
        <TouchableOpacity
          className="bg-amethyst-500 rounded-xl py-3 items-center mt-4"
          activeOpacity={0.8}
          onPress={() => setSelectedListing(item)}
        >
          <Text className="text-white font-semibold text-sm">Buy</Text>
        </TouchableOpacity>
      </View>
    );
  };

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      {/* Header */}
      <View className="px-5 pt-3 pb-4">
        <Text className="text-2xl font-bold text-slate-900">Buy USDT</Text>
        <Text className="text-sm text-slate-500 mt-0.5">
          Secure P2P exchange powered by Swipe
        </Text>
      </View>

      {/* Loading state */}
      {listings === undefined && (
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator size="large" color="#9333ea" />
          <Text className="text-slate-500 mt-3">Loading listings...</Text>
        </View>
      )}

      {/* Empty state */}
      {listings !== undefined && listings.length === 0 && (
        <View className="flex-1 items-center justify-center px-8">
          <View className="h-16 w-16 rounded-full bg-slate-100 items-center justify-center mb-4">
            <Ionicons name="cash-outline" size={32} color="#94a3b8" />
          </View>
          <Text className="text-lg font-semibold text-slate-700">
            No listings available
          </Text>
          <Text className="text-sm text-slate-500 text-center mt-2 max-w-[260px]">
            Check back soon — sellers are adding USDT listings
          </Text>
        </View>
      )}

      {/* Listings */}
      {listings !== undefined && listings.length > 0 && (
        <FlatList
          data={listings}
          keyExtractor={(item) => item._id}
          renderItem={renderListing}
          contentContainerStyle={{ paddingBottom: 20 }}
        />
      )}

      {/* Wallet Input Bottom-Sheet Modal */}
      <WalletModal
        visible={!!selectedListing}
        onClose={() => setSelectedListing(null)}
        onSubmit={async (wallet) => {
          if (selectedListing) await processPurchase(selectedListing, wallet);
          setSelectedListing(null);
        }}
      />
    </SafeAreaView>
  );
}

/* ------------------------------------------------------------------ */
/*  Bottom-sheet wallet modal                                          */
/* ------------------------------------------------------------------ */

function WalletModal({
  visible,
  onClose,
  onSubmit,
}: {
  visible: boolean;
  onClose: () => void;
  onSubmit: (wallet: string) => void;
}) {
  const [wallet, setWallet] = useState("");
  const isValid = wallet.startsWith("T") && wallet.length === 34;

  return (
    <Modal visible={visible} animationType="slide" transparent>
      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        className="flex-1 justify-end"
      >
        {/* Backdrop */}
        <TouchableOpacity
          className="flex-1 bg-black/40"
          activeOpacity={1}
          onPress={onClose}
        />

        {/* Sheet */}
        <View className="bg-white rounded-t-3xl px-5 pt-5 pb-8">
          {/* Title row */}
          <View className="flex-row items-center justify-between mb-4">
            <Text className="text-lg font-bold text-slate-900">
              Enter TRC20 Wallet
            </Text>
            <TouchableOpacity onPress={onClose}>
              <Ionicons name="close" size={24} color="#64748b" />
            </TouchableOpacity>
          </View>

          {/* Description */}
          <Text className="text-sm text-slate-500 mb-4">
            Your TRON wallet address to receive USDT. Must start with T, 34
            characters.
          </Text>

          {/* Wallet input */}
          <TextInput
            className={`border rounded-xl px-4 py-3.5 text-base text-slate-900 mb-1 ${
              wallet.length > 0 && !isValid
                ? "border-red-400"
                : "border-slate-200"
            }`}
            placeholder="T..."
            placeholderTextColor="#94a3b8"
            value={wallet}
            onChangeText={setWallet}
            autoCapitalize="characters"
            maxLength={34}
            autoFocus
          />

          {/* Character count */}
          <Text
            className={`text-xs mb-5 ${
              wallet.length > 0 && !isValid ? "text-red-500" : "text-slate-400"
            }`}
          >
            {wallet.length}/34 characters
          </Text>

          {/* Action buttons */}
          <View className="flex-row gap-3">
            <TouchableOpacity
              className="flex-1 border border-slate-200 rounded-xl py-3.5 items-center"
              onPress={() => {
                setWallet("");
                onClose();
              }}
            >
              <Text className="text-slate-700 font-semibold">Cancel</Text>
            </TouchableOpacity>
            <TouchableOpacity
              className={`flex-1 rounded-xl py-3.5 items-center ${
                isValid ? "bg-amethyst-500" : "bg-slate-200"
              }`}
              disabled={!isValid}
              onPress={() => {
                onSubmit(wallet);
                setWallet("");
              }}
            >
              <Text
                className={`font-semibold ${
                  isValid ? "text-white" : "text-slate-400"
                }`}
              >
                Continue
              </Text>
            </TouchableOpacity>
          </View>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}
