import { useState } from "react";
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import { useRouter, useLocalSearchParams } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import type { Id } from "../../convex/_generated/dataModel";

type Order = {
  _id: Id<"orders">;
  customerName: string;
  customerPhone: string;
  deliveryLocation: string;
  deliveryAddress: string;
  quantity: number;
  unitPrice: number;
  totalAmount: number;
  currency: string;
  status: string;
  paymentStatus: string;
  createdAt: number;
  updatedAt: number;
  productId: Id<"products">;
  variantId: Id<"productVariants">;
};

export default function SellerOrdersScreen() {
  const router = useRouter();
  const { sessionToken } = useLocalSearchParams<{ sessionToken: string }>();
  const [selectedOrderId, setSelectedOrderId] = useState<Id<"orders"> | null>(
    null
  );

  const orders = useQuery(
    api.orders.listByStore,
    sessionToken ? { sessionToken } : "skip"
  );
  const fulfillOrder = useMutation(api.orders.fulfillOrder);

  if (!sessionToken) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center">
        <Text className="text-slate-500">No session. Please log in.</Text>
        <TouchableOpacity
          className="mt-4 bg-purple-600 rounded-xl px-6 py-3"
          onPress={() => router.back()}
        >
          <Text className="text-white font-semibold">Go Back</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  const handleStatusUpdate = (order: Order, newStatus: string) => {
    const label = newStatus === "shipped" ? "Mark as Shipped" : "Mark as Delivered";
    Alert.alert(
      label,
      `Update order for ${order.customerName} to "${newStatus}"?`,
      [
        { text: "Cancel", style: "cancel" },
        {
          text: "Confirm",
          onPress: async () => {
            try {
              const result = await fulfillOrder({
                sessionToken: sessionToken!,
                orderId: order._id,
                newStatus,
              });
              if (result.success) {
                Alert.alert("Updated", `Order marked as ${newStatus}.`);
              } else {
                Alert.alert("Error", `Cannot update: ${"reason" in result ? result.reason : "unknown"}`);
              }
            } catch (err: any) {
              Alert.alert("Error", err.message || "Failed to update order");
            }
          },
        },
      ]
    );
  };

  const renderOrder = ({ item }: { item: Order }) => {
    const isExpanded = selectedOrderId === item._id;

    return (
      <TouchableOpacity
        className="mx-4 mb-3 bg-white rounded-2xl border border-slate-100 overflow-hidden"
        activeOpacity={0.7}
        onPress={() => setSelectedOrderId(isExpanded ? null : item._id)}
      >
        {/* Summary Row */}
        <View className="p-4">
          <View className="flex-row items-start justify-between">
            <View className="flex-1">
              <Text className="text-base font-semibold text-slate-900">
                {item.customerName}
              </Text>
              <Text className="text-xs text-slate-500 mt-0.5">
                {new Date(item.createdAt).toLocaleDateString("en-US", {
                  month: "short",
                  day: "numeric",
                  year: "numeric",
                  hour: "2-digit",
                  minute: "2-digit",
                })}
              </Text>
            </View>
            <View className="items-end">
              <Text className="text-base font-bold text-slate-900">
                {item.totalAmount.toFixed(2)} {item.currency}
              </Text>
              <StatusBadge status={item.status} />
            </View>
          </View>

          {/* Location preview */}
          <View className="flex-row items-center mt-2">
            <Ionicons name="location-outline" size={14} color="#94a3b8" />
            <Text className="text-xs text-slate-500 ml-1" numberOfLines={1}>
              {item.deliveryLocation}
            </Text>
          </View>
        </View>

        {/* Expanded Details */}
        {isExpanded && (
          <OrderDetails
            order={item}
            sessionToken={sessionToken!}
            onStatusUpdate={handleStatusUpdate}
          />
        )}
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      {/* Header */}
      <View className="px-4 pt-2 pb-4 flex-row items-center">
        <TouchableOpacity className="mr-3 p-1" onPress={() => router.back()}>
          <Ionicons name="arrow-back" size={24} color="#1e293b" />
        </TouchableOpacity>
        <View className="flex-1">
          <Text className="text-2xl font-bold text-slate-900">Orders</Text>
          <Text className="text-sm text-slate-500">
            {orders ? `${orders.length} orders` : "Loading..."}
          </Text>
        </View>
      </View>

      {/* Loading */}
      {orders === undefined && (
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator size="large" color="#9333ea" />
          <Text className="text-slate-500 mt-3">Loading orders...</Text>
        </View>
      )}

      {/* Empty */}
      {orders !== undefined && orders.length === 0 && (
        <View className="flex-1 items-center justify-center px-8">
          <Ionicons name="receipt-outline" size={64} color="#cbd5e1" />
          <Text className="text-lg font-semibold text-slate-700 mt-4">
            No Orders
          </Text>
          <Text className="text-sm text-slate-500 text-center mt-2">
            Orders will appear here once customers make purchases.
          </Text>
        </View>
      )}

      {/* Order List */}
      {orders !== undefined && orders.length > 0 && (
        <FlatList
          data={orders as Order[]}
          keyExtractor={(item) => item._id}
          renderItem={renderOrder}
          contentContainerStyle={{ paddingBottom: 20 }}
        />
      )}
    </SafeAreaView>
  );
}

/** Expanded order details panel */
function OrderDetails({
  order,
  sessionToken,
  onStatusUpdate,
}: {
  order: Order;
  sessionToken: string;
  onStatusUpdate: (order: Order, newStatus: string) => void;
}) {
  const orderDetails = useQuery(api.orders.getOrderWithDetails, {
    orderId: order._id,
  });

  return (
    <View className="px-4 pb-4 pt-2 border-t border-slate-100 bg-slate-50">
      {/* Product Info */}
      {orderDetails?.product && (
        <View className="mb-3">
          <Text className="text-xs text-slate-500 uppercase tracking-wide mb-1">
            Product
          </Text>
          <Text className="text-sm font-medium text-slate-900">
            {orderDetails.product.name}
          </Text>
          {orderDetails.variant && (
            <Text className="text-xs text-slate-500 mt-0.5">
              Variant: {orderDetails.variant.variantName}
            </Text>
          )}
        </View>
      )}

      {/* Order Details */}
      <View className="flex-row mb-3">
        <View className="flex-1">
          <Text className="text-xs text-slate-500">Qty</Text>
          <Text className="text-sm font-medium text-slate-900">
            {order.quantity}
          </Text>
        </View>
        <View className="flex-1">
          <Text className="text-xs text-slate-500">Unit Price</Text>
          <Text className="text-sm font-medium text-slate-900">
            {order.unitPrice.toFixed(2)} {order.currency}
          </Text>
        </View>
        <View className="flex-1">
          <Text className="text-xs text-slate-500">Payment</Text>
          <Text
            className={`text-sm font-medium ${
              order.paymentStatus === "paid"
                ? "text-green-700"
                : "text-yellow-700"
            }`}
          >
            {order.paymentStatus}
          </Text>
        </View>
      </View>

      {/* Customer Info */}
      <View className="mb-3">
        <Text className="text-xs text-slate-500 uppercase tracking-wide mb-1">
          Delivery
        </Text>
        <Text className="text-sm text-slate-900">{order.deliveryAddress}</Text>
        <Text className="text-xs text-slate-500 mt-0.5">
          Phone: {order.customerPhone}
        </Text>
      </View>

      {/* Action Buttons */}
      {order.status === "paid" && (
        <TouchableOpacity
          className="bg-blue-600 rounded-xl py-3 items-center"
          activeOpacity={0.8}
          onPress={() => onStatusUpdate(order, "shipped")}
        >
          <Text className="text-white font-semibold">Mark as Shipped</Text>
        </TouchableOpacity>
      )}
      {order.status === "shipped" && (
        <TouchableOpacity
          className="bg-slate-700 rounded-xl py-3 items-center"
          activeOpacity={0.8}
          onPress={() => onStatusUpdate(order, "delivered")}
        >
          <Text className="text-white font-semibold">Mark as Delivered</Text>
        </TouchableOpacity>
      )}
    </View>
  );
}

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { bg: string; text: string }> = {
    pending: { bg: "bg-yellow-100", text: "text-yellow-700" },
    paid: { bg: "bg-green-100", text: "text-green-700" },
    shipped: { bg: "bg-blue-100", text: "text-blue-700" },
    delivered: { bg: "bg-slate-100", text: "text-slate-600" },
    cancelled: { bg: "bg-red-100", text: "text-red-700" },
  };
  const { bg, text } = config[status] ?? {
    bg: "bg-slate-100",
    text: "text-slate-600",
  };

  return (
    <View className={`${bg} px-2 py-0.5 rounded-full mt-1`}>
      <Text className={`text-xs font-medium ${text} capitalize`}>{status}</Text>
    </View>
  );
}
