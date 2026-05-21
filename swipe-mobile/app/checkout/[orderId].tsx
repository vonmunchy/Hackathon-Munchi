import { useState } from "react";
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  Alert,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useRouter, useLocalSearchParams } from "expo-router";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import type { Id } from "../../convex/_generated/dataModel";
import MvrAmount from "../../components/MvrAmount";
import LoadingSpinner from "../../components/LoadingSpinner";
import { Ionicons } from "@expo/vector-icons";

export default function CheckoutPage() {
  const router = useRouter();
  const { orderId, token } = useLocalSearchParams<{
    orderId: string;
    token: string;
  }>();

  const order = useQuery(
    api.orders.getByIdWithToken,
    orderId && token
      ? { orderId: orderId as Id<"orders">, accessToken: token }
      : "skip"
  );

  const confirmPayment = useMutation(api.orders.confirmPaymentPublic);
  const [confirming, setConfirming] = useState(false);

  if (order === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50">
        <LoadingSpinner message="Loading order..." />
      </SafeAreaView>
    );
  }

  if (order === null) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center px-8">
        <Ionicons name="alert-circle-outline" size={48} color="#e11d48" />
        <Text className="text-lg font-semibold text-slate-700 mt-3">Order not found</Text>
        <Text className="text-sm text-slate-400 mt-1 text-center">
          This order may have expired or the link is invalid.
        </Text>
        <TouchableOpacity onPress={() => router.replace("/")} className="mt-5">
          <Text className="text-amethyst-600 font-medium">Back to Marketplace</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  const isPaid = order.paymentStatus === "paid";
  const isCancelled = order.paymentStatus === "cancelled";

  const handleSimulatePayment = async () => {
    setConfirming(true);
    try {
      const result = await confirmPayment({
        orderId: order._id,
      });
      if (!result.success) {
        Alert.alert("Payment Failed", result.reason ?? "Could not process payment.");
      }
    } catch (err: any) {
      Alert.alert("Error", err.message ?? "Something went wrong.");
    } finally {
      setConfirming(false);
    }
  };

  const statusColor = isPaid
    ? "bg-emerald-100 text-emerald-700"
    : isCancelled
      ? "bg-ruby-100 text-ruby-600"
      : "bg-amber-100 text-amber-700";

  const statusLabel = isPaid
    ? "Paid"
    : isCancelled
      ? "Cancelled"
      : "Pending Payment";

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      {/* Header */}
      <View className="flex-row items-center px-4 py-3 bg-white border-b border-slate-100">
        <TouchableOpacity onPress={() => router.replace("/")} className="mr-3">
          <Ionicons name="arrow-back" size={24} color="#1e293b" />
        </TouchableOpacity>
        <Text className="text-lg font-semibold text-slate-900">Checkout</Text>
      </View>

      <ScrollView className="flex-1" contentContainerStyle={{ padding: 16 }}>
        {/* Success state */}
        {isPaid ? (
          <View className="items-center py-8 mb-4">
            <View className="w-20 h-20 rounded-full bg-emerald-100 items-center justify-center mb-4">
              <Ionicons name="checkmark-circle" size={48} color="#059669" />
            </View>
            <Text className="text-xl font-bold text-slate-900">Payment Successful!</Text>
            <Text className="text-sm text-slate-500 mt-1">
              Your order has been confirmed.
            </Text>
          </View>
        ) : null}

        {/* Order summary card */}
        <View className="bg-white rounded-2xl p-4 mb-4">
          <Text className="text-sm font-semibold text-slate-500 mb-3">ORDER SUMMARY</Text>

          <View className="flex-row justify-between mb-2">
            <Text className="text-sm text-slate-500">Quantity</Text>
            <Text className="text-sm font-medium text-slate-800">{order.quantity}</Text>
          </View>

          <View className="flex-row justify-between mb-2">
            <Text className="text-sm text-slate-500">Unit Price</Text>
            <MvrAmount
              amount={order.unitPrice}
              className="text-sm font-medium text-slate-800"
            />
          </View>

          <View className="h-px bg-slate-100 my-2" />

          <View className="flex-row justify-between">
            <Text className="text-base font-semibold text-slate-900">Total</Text>
            <MvrAmount
              amount={order.totalAmount}
              className="text-base font-bold text-amethyst-600"
            />
          </View>
        </View>

        {/* Payment status */}
        <View className="bg-white rounded-2xl p-4 mb-4">
          <Text className="text-sm font-semibold text-slate-500 mb-3">PAYMENT STATUS</Text>
          <View className={`self-start px-3 py-1.5 rounded-full ${statusColor.split(" ")[0]}`}>
            <Text className={`text-sm font-semibold ${statusColor.split(" ")[1]}`}>
              {statusLabel}
            </Text>
          </View>
        </View>

        {/* Delivery info */}
        <View className="bg-white rounded-2xl p-4 mb-4">
          <Text className="text-sm font-semibold text-slate-500 mb-3">DELIVERY</Text>
          <Text className="text-sm text-slate-800">{order.customerName}</Text>
          <Text className="text-sm text-slate-500 mt-1">{order.customerPhone}</Text>
          <Text className="text-sm text-slate-500 mt-1">
            {order.deliveryLocation} - {order.deliveryAddress}
          </Text>
        </View>

        {/* USDT QR data if available */}
        {order.swipeQrData ? (
          <View className="bg-white rounded-2xl p-4 mb-4 items-center">
            <Text className="text-sm font-semibold text-slate-500 mb-2">PAYMENT QR</Text>
            <View className="bg-slate-100 rounded-xl p-4">
              <Text className="text-xs text-slate-600 font-mono text-center">
                {order.swipeQrData}
              </Text>
            </View>
            {order.swipeShortCode ? (
              <Text className="text-sm text-slate-600 mt-2">
                Code: <Text className="font-bold">{order.swipeShortCode}</Text>
              </Text>
            ) : null}
          </View>
        ) : null}

        {/* Simulate payment button (demo) */}
        {!isPaid && !isCancelled ? (
          <TouchableOpacity
            onPress={handleSimulatePayment}
            disabled={confirming}
            className={`py-4 rounded-2xl items-center mt-2 ${
              confirming ? "bg-amethyst-400" : "bg-amethyst-600"
            }`}
          >
            <Text className="text-base font-bold text-white">
              {confirming ? "Processing..." : "Simulate Payment (Demo)"}
            </Text>
          </TouchableOpacity>
        ) : null}

        {/* Back to marketplace */}
        {isPaid || isCancelled ? (
          <TouchableOpacity
            onPress={() => router.replace("/")}
            className="py-4 rounded-2xl items-center mt-2 bg-amethyst-600"
          >
            <Text className="text-base font-bold text-white">Back to Marketplace</Text>
          </TouchableOpacity>
        ) : null}
      </ScrollView>
    </SafeAreaView>
  );
}
