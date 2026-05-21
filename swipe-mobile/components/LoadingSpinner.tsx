import { View, Text, ActivityIndicator } from "react-native";

interface LoadingSpinnerProps {
  message?: string;
  className?: string;
}

export default function LoadingSpinner({ message, className }: LoadingSpinnerProps) {
  return (
    <View className={`flex-1 items-center justify-center ${className ?? ""}`}>
      <ActivityIndicator size="large" color="#9333ea" />
      {message ? (
        <Text className="mt-3 text-sm text-slate-500">{message}</Text>
      ) : null}
    </View>
  );
}
