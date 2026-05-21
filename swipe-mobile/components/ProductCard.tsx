import { View, Text, TouchableOpacity } from "react-native";
import { Image } from "expo-image";
import MvrAmount from "./MvrAmount";
import { Ionicons } from "@expo/vector-icons";

interface ProductCardProps {
  name: string;
  imageUrl: string;
  price: number;
  storeName: string;
  category?: string | null;
  onPress: () => void;
}

export default function ProductCard({
  name,
  imageUrl,
  price,
  storeName,
  category,
  onPress,
}: ProductCardProps) {
  return (
    <TouchableOpacity
      onPress={onPress}
      activeOpacity={0.7}
      className="flex-1 bg-white rounded-2xl overflow-hidden shadow-sm border border-slate-100"
    >
      {/* Image */}
      <View className="relative">
        <Image
          source={{ uri: imageUrl }}
          className="w-full aspect-square"
          contentFit="cover"
          transition={200}
        />
        {category ? (
          <View className="absolute top-2 left-2 bg-white/90 px-2 py-0.5 rounded-full">
            <Text className="text-[10px] font-medium text-slate-700">{category}</Text>
          </View>
        ) : null}
      </View>

      {/* Info */}
      <View className="p-3">
        <Text
          className="text-sm font-medium text-slate-800 leading-tight"
          numberOfLines={2}
          ellipsizeMode="tail"
        >
          {name}
        </Text>
        <View className="flex-row items-center mt-1.5 gap-1">
          <Ionicons name="storefront-outline" size={11} color="#94a3b8" />
          <Text className="text-xs text-slate-400" numberOfLines={1}>{storeName}</Text>
        </View>
        <MvrAmount amount={price} className="text-sm font-bold text-amethyst-600 mt-1.5" />
      </View>
    </TouchableOpacity>
  );
}
