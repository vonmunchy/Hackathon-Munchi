import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import { Image } from "expo-image";
import { useQuery } from "convex/react";
import { api } from "../../convex/_generated/api";
import { useRouter, useLocalSearchParams } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import type { Id } from "../../convex/_generated/dataModel";

type Product = {
  _id: Id<"products">;
  name: string;
  description?: string;
  basePrice: number;
  currency: string;
  category?: string;
  imageUrls: string[];
  status: string;
  createdAt: number;
};

export default function SellerProductsScreen() {
  const router = useRouter();
  const { sessionToken, storeId } = useLocalSearchParams<{
    sessionToken: string;
    storeId: string;
  }>();

  const products = useQuery(
    api.products.listByStore,
    storeId ? { storeId: storeId as Id<"stores"> } : "skip"
  );

  if (!storeId) {
    return (
      <SafeAreaView className="flex-1 bg-slate-50 items-center justify-center">
        <Text className="text-slate-500">No store selected.</Text>
        <TouchableOpacity
          className="mt-4 bg-purple-600 rounded-xl px-6 py-3"
          onPress={() => router.back()}
        >
          <Text className="text-white font-semibold">Go Back</Text>
        </TouchableOpacity>
      </SafeAreaView>
    );
  }

  const renderProduct = ({ item }: { item: Product }) => {
    const hasImage = item.imageUrls.length > 0;
    const statusColor =
      item.status === "active"
        ? "bg-green-100 text-green-700"
        : item.status === "archived"
          ? "bg-slate-100 text-slate-600"
          : "bg-yellow-100 text-yellow-700";
    const [statusBg, statusText] = statusColor.split(" ");

    return (
      <View className="mx-4 mb-3 bg-white rounded-2xl border border-slate-100 overflow-hidden">
        <View className="flex-row">
          {/* Product Image */}
          <View className="w-24 h-24 bg-slate-100 items-center justify-center">
            {hasImage ? (
              <Image
                source={{ uri: item.imageUrls[0] }}
                style={{ width: 96, height: 96 }}
                contentFit="cover"
              />
            ) : (
              <Ionicons name="image-outline" size={32} color="#cbd5e1" />
            )}
          </View>

          {/* Product Info */}
          <View className="flex-1 p-3 justify-between">
            <View>
              <View className="flex-row items-center justify-between">
                <Text
                  className="text-base font-semibold text-slate-900 flex-1"
                  numberOfLines={1}
                >
                  {item.name}
                </Text>
                <View className={`${statusBg} px-2 py-0.5 rounded-full ml-2`}>
                  <Text className={`text-xs font-medium ${statusText} capitalize`}>
                    {item.status}
                  </Text>
                </View>
              </View>
              {item.category ? (
                <Text className="text-xs text-slate-500 mt-0.5">
                  {item.category}
                </Text>
              ) : null}
            </View>
            <View className="flex-row items-center justify-between mt-2">
              <Text className="text-lg font-bold text-purple-700">
                {item.basePrice.toFixed(2)}{" "}
                <Text className="text-xs font-normal text-slate-500">
                  {item.currency}
                </Text>
              </Text>
            </View>
          </View>
        </View>

        {/* Variants (loaded inline) */}
        <ProductVariants productId={item._id} />
      </View>
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
          <Text className="text-2xl font-bold text-slate-900">Products</Text>
          <Text className="text-sm text-slate-500">
            {products ? `${products.length} products` : "Loading..."}
          </Text>
        </View>
      </View>

      {/* Loading */}
      {products === undefined && (
        <View className="flex-1 items-center justify-center">
          <ActivityIndicator size="large" color="#9333ea" />
          <Text className="text-slate-500 mt-3">Loading products...</Text>
        </View>
      )}

      {/* Empty */}
      {products !== undefined && products.length === 0 && (
        <View className="flex-1 items-center justify-center px-8">
          <Ionicons name="cube-outline" size={64} color="#cbd5e1" />
          <Text className="text-lg font-semibold text-slate-700 mt-4">
            No Products
          </Text>
          <Text className="text-sm text-slate-500 text-center mt-2">
            You haven't added any products yet.
          </Text>
        </View>
      )}

      {/* Product List */}
      {products !== undefined && products.length > 0 && (
        <FlatList
          data={products}
          keyExtractor={(item) => item._id}
          renderItem={renderProduct}
          contentContainerStyle={{ paddingBottom: 20 }}
        />
      )}
    </SafeAreaView>
  );
}

/** Shows variant stock counts for a product */
function ProductVariants({ productId }: { productId: Id<"products"> }) {
  const product = useQuery(api.products.getWithVariants, { productId });
  const variants = product?.variants;

  if (!variants || variants.length === 0) return null;

  return (
    <View className="px-3 pb-3 pt-1 flex-row flex-wrap gap-2 border-t border-slate-50">
      {variants.map((v) => (
        <View
          key={v._id}
          className={`flex-row items-center px-2 py-1 rounded-lg ${
            v.stockAvailable <= 3 ? "bg-red-50" : "bg-slate-50"
          }`}
        >
          <Text className="text-xs text-slate-700 font-medium">
            {v.variantName}
          </Text>
          <Text
            className={`text-xs ml-1 font-bold ${
              v.stockAvailable <= 3 ? "text-red-600" : "text-slate-600"
            }`}
          >
            ({v.stockAvailable})
          </Text>
        </View>
      ))}
    </View>
  );
}
