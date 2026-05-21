import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { useQuery, useMutation } from "convex/react";
import { api } from "../../convex/_generated/api";
import { useRouter, useLocalSearchParams } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

export default function SellerDashboardScreen() {
  const router = useRouter();
  const { sessionToken } = useLocalSearchParams<{ sessionToken: string }>();

  const session = useQuery(
    api.sessions.validate,
    sessionToken ? { token: sessionToken } : "skip"
  );
  const stats = useQuery(
    api.orders.getDashboardStats,
    sessionToken ? { sessionToken } : "skip"
  );
  const orders = useQuery(
    api.orders.listByStore,
    sessionToken ? { sessionToken } : "skip"
  );

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

  if (session === undefined || stats === undefined) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center">
        <ActivityIndicator size="large" color="#9333ea" />
        <Text className="text-slate-500 mt-3">Loading dashboard...</Text>
      </SafeAreaView>
    );
  }

  if (!session) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center">
        <Ionicons name="alert-circle-outline" size={48} color="#ef4444" />
        <Text className="text-slate-700 font-semibold mt-3">Session Expired</Text>
        <TouchableOpacity
          className="mt-4 bg-purple-600 rounded-xl px-6 py-3"
          onPress={() => router.back()}
        >
          <Text className="text-white font-semibold">Log In Again</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  const recentOrders = orders?.slice(0, 10) ?? [];

  return (
    <SafeAreaView className="flex-1 bg-slate-50" edges={["top"]}>
      <ScrollView className="flex-1" contentContainerStyle={{ paddingBottom: 30 }}>
        {/* Header */}
        <View className="px-4 pt-2 pb-4 flex-row items-center">
          <TouchableOpacity
            className="mr-3 p-1"
            onPress={() => router.back()}
          >
            <Ionicons name="arrow-back" size={24} color="#1e293b" />
          </TouchableOpacity>
          <View className="flex-1">
            <Text className="text-2xl font-bold text-slate-900">
              {session.storeName}
            </Text>
            <Text className="text-sm text-slate-500">Dashboard</Text>
          </View>
        </View>

        {/* Stats Grid */}
        <View className="px-4">
          <Text className="text-lg font-bold text-slate-900 mb-3">
            Overview
          </Text>
          <View className="flex-row flex-wrap gap-3">
            <DashStatCard
              label="Today's Sales"
              value={`${(stats?.totalSalesToday ?? 0).toLocaleString()} MVR`}
              icon="trending-up"
              bgColor="bg-purple-50"
              iconColor="#9333ea"
            />
            <DashStatCard
              label="Total Sales"
              value={`${(stats?.totalSalesAllTime ?? 0).toLocaleString()} MVR`}
              icon="cash-outline"
              bgColor="bg-green-50"
              iconColor="#15803d"
            />
            <DashStatCard
              label="Orders Paid"
              value={String(stats?.paidCount ?? 0)}
              icon="checkmark-circle-outline"
              bgColor="bg-blue-50"
              iconColor="#2563eb"
            />
            <DashStatCard
              label="Pending Orders"
              value={String(stats?.pendingCount ?? 0)}
              icon="hourglass-outline"
              bgColor="bg-yellow-50"
              iconColor="#ca8a04"
            />
            <DashStatCard
              label="Awaiting Shipment"
              value={String(stats?.awaitingShipmentCount ?? 0)}
              icon="airplane-outline"
              bgColor="bg-orange-50"
              iconColor="#ea580c"
            />
            <DashStatCard
              label="Low Stock"
              value={String(stats?.lowStockCount ?? 0)}
              icon="alert-circle-outline"
              bgColor="bg-red-50"
              iconColor="#dc2626"
            />
          </View>
        </View>

        {/* Quick Actions */}
        <View className="px-4 mt-6 flex-row gap-3">
          <TouchableOpacity
            className="flex-1 bg-purple-600 rounded-xl py-4 items-center flex-row justify-center gap-2"
            activeOpacity={0.8}
            onPress={() =>
              router.push({
                pathname: "/seller/products",
                params: { sessionToken, storeId: session.storeId },
              })
            }
          >
            <Ionicons name="cube-outline" size={20} color="white" />
            <Text className="text-white font-semibold">Products</Text>
          </TouchableOpacity>
          <TouchableOpacity
            className="flex-1 bg-slate-800 rounded-xl py-4 items-center flex-row justify-center gap-2"
            activeOpacity={0.8}
            onPress={() =>
              router.push({
                pathname: "/seller/orders",
                params: { sessionToken },
              })
            }
          >
            <Ionicons name="receipt-outline" size={20} color="white" />
            <Text className="text-white font-semibold">Orders</Text>
          </TouchableOpacity>
        </View>

        {/* Recent Orders */}
        <View className="px-4 mt-6">
          <View className="flex-row items-center justify-between mb-3">
            <Text className="text-lg font-bold text-slate-900">
              Recent Orders
            </Text>
            <TouchableOpacity
              onPress={() =>
                router.push({
                  pathname: "/seller/orders",
                  params: { sessionToken },
                })
              }
            >
              <Text className="text-purple-600 text-sm font-medium">
                View All
              </Text>
            </TouchableOpacity>
          </View>

          {recentOrders.length === 0 ? (
            <View className="bg-white rounded-xl p-8 items-center border border-slate-100">
              <Ionicons name="receipt-outline" size={40} color="#cbd5e1" />
              <Text className="text-slate-500 mt-3 text-center">
                No orders yet. Share your store to start selling!
              </Text>
            </View>
          ) : (
            recentOrders.map((order) => (
              <View
                key={order._id}
                className="bg-white rounded-xl p-4 mb-2 border border-slate-100"
              >
                <View className="flex-row items-center justify-between">
                  <View className="flex-1">
                    <Text className="text-sm font-semibold text-slate-900">
                      {order.customerName}
                    </Text>
                    <Text className="text-xs text-slate-500 mt-0.5">
                      {order.deliveryLocation} - {new Date(order.createdAt).toLocaleDateString()}
                    </Text>
                  </View>
                  <View className="items-end">
                    <Text className="text-sm font-bold text-slate-900">
                      {order.totalAmount.toFixed(2)} {order.currency}
                    </Text>
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

function DashStatCard({
  label,
  value,
  icon,
  bgColor,
  iconColor,
}: {
  label: string;
  value: string;
  icon: keyof typeof Ionicons.glyphMap;
  bgColor: string;
  iconColor: string;
}) {
  return (
    <View
      className={`${bgColor} rounded-xl p-4 border border-slate-100`}
      style={{ width: "47%" }}
    >
      <Ionicons name={icon} size={22} color={iconColor} />
      <Text className="text-xl font-bold text-slate-900 mt-2">{value}</Text>
      <Text className="text-xs text-slate-500 mt-1">{label}</Text>
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
  const { bg, text } = config[status] ?? { bg: "bg-slate-100", text: "text-slate-600" };

  return (
    <View className={`${bg} px-2 py-0.5 rounded-full mt-1`}>
      <Text className={`text-xs font-medium ${text} capitalize`}>{status}</Text>
    </View>
  );
}
