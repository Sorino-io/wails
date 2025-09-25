<template>
  <div>
    <div class="mb-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-2xl font-bold text-gray-900 mb-2">
            {{ $t("products.title") }}
          </h1>
          <p class="text-gray-600">
            {{ $t("products.list") }}
          </p>
        </div>
        <button @click="showCreateModal = true" class="btn btn-primary">
          {{ $t("products.create") }}
        </button>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="mb-6">
      <div class="max-w-md">
        <input
          v-model="searchQuery"
          type="text"
          :placeholder="$t('products.search_placeholder')"
          class="form-input"
          @input="debouncedSearch"
        />
      </div>
    </div>

    <!-- Products Table -->
    <div class="card">
      <div class="overflow-x-auto">
        <table class="table">
          <thead>
            <tr>
              <th>{{ $t("fields.name") }}</th>
              <th>{{ $t("fields.description") }}</th>
              <th>{{ $t("fields.price") }}</th>
              <th>{{ $t("fields.sku") }}</th>
              <th>{{ $t("fields.status") }}</th>
              <th class="text-center">{{ $t("fields.actions") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="text-center py-8">
                <div class="spinner mx-auto"></div>
                <p class="mt-2 text-gray-500">{{ $t("messages.loading") }}</p>
              </td>
            </tr>
            <tr v-else-if="products.length === 0">
              <td colspan="6" class="text-center py-8">
                <p class="text-gray-500">{{ $t("messages.no_data") }}</p>
              </td>
            </tr>
            <tr v-else v-for="product in products" :key="product.id">
              <td class="font-medium">{{ product.name }}</td>
              <td class="max-w-xs truncate">
                {{ product.description || "---" }}
              </td>
              <td class="font-medium">
                {{ formatPrice(product.unit_price_cents) }}
              </td>
              <td>{{ product.sku || "---" }}</td>
              <td>
                <span
                  :class="
                    product.active
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  "
                  class="px-2 py-1 rounded-full text-xs font-medium"
                >
                  {{
                    product.active ? $t("status.active") : $t("status.inactive")
                  }}
                </span>
              </td>
              <td class="text-center">
                <div class="inline-flex items-center justify-center gap-2">
                  <button
                    @click="editProduct(product)"
                    class="text-blue-600 hover:text-blue-900 p-1 rounded"
                    :title="$t('common.edit')"
                  >
                    <PencilIcon class="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="mt-6 flex items-center justify-between">
      <div class="text-sm text-gray-700">
        {{ $t("pagination.showing") }} {{ (currentPage - 1) * pageSize + 1 }}
        {{ $t("pagination.to") }}
        {{ Math.min(currentPage * pageSize, totalCount) }}
        {{ $t("pagination.of") }} {{ totalCount }}
        {{ $t("pagination.results") }}
      </div>
      <div class="flex space-x-2 space-x-reverse">
        <button
          @click="previousPage"
          :disabled="currentPage === 1"
          class="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ $t("pagination.previous") }}
        </button>
        <button
          @click="nextPage"
          :disabled="currentPage === totalPages"
          class="btn btn-secondary disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {{ $t("pagination.next") }}
        </button>
      </div>
    </div>

    <!-- Create/Edit Modal -->
    <div
      v-if="showCreateModal || showEditModal"
      class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50"
    >
      <div
        class="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white"
        @click.stop
      >
        <div class="mt-3">
          <h3 class="text-lg font-medium text-gray-900 mb-4">
            {{ showCreateModal ? $t("products.create") : $t("products.edit") }}
          </h3>

          <!-- Error Message -->
          <div v-if="errorMessage" class="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
            {{ errorMessage }}
          </div>

          <!-- Success Message -->
          <div v-if="successMessage" class="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
            {{ successMessage }}
          </div>

          <form @submit.prevent="saveProduct" class="space-y-4">
            <div>
              <label class="form-label">{{ $t("fields.name") }} *</label>
              <input
                v-model="currentProduct.name"
                type="text"
                required
                class="form-input"
              />
            </div>

            <div>
              <label class="form-label">{{ $t("fields.description") }}</label>
              <textarea
                v-model="currentProduct.description"
                rows="3"
                class="form-input"
              ></textarea>
            </div>

            <div>
              <label class="form-label">{{ $t("fields.price") }} *</label>
              <input
                v-model.number="currentProduct.price"
                type="number"
                step="0.01"
                min="0"
                required
                class="form-input"
              />
            </div>

            <div>
              <label class="form-label">{{ $t("fields.sku") }}</label>
              <input
                v-model="currentProduct.sku"
                type="text"
                class="form-input"
              />
            </div>

            <div class="flex justify-between items-center pt-4">
              <div v-if="showEditModal">
                <button type="button" @click="confirmDeleteProduct" class="btn btn-danger">{{ $t('actions.delete') }}</button>
              </div>
              <div class="flex space-x-3 space-x-reverse">
              <button
                type="button"
                @click="closeModal"
                class="btn btn-secondary"
              >
                {{ $t("actions.cancel") }}
              </button>
              <button
                type="submit"
                :disabled="loading || currentProduct.price <= 0"
                class="btn btn-primary disabled:opacity-50"
              >
                {{ loading ? $t("messages.loading") : $t("actions.save") }}
              </button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
    <!-- Delete Product Confirm -->
    <div v-if="showDeleteProductConfirm" class="fixed inset-0 bg-gray-700 bg-opacity-60 flex items-center justify-center z-[70]">
      <div class="bg-white rounded-md shadow p-6 w-full max-w-sm">
        <h3 class="text-lg font-semibold mb-4">{{ $t('products.delete_title') }}</h3>
        <p class="text-sm text-gray-600 mb-6">{{ $t('products.delete_confirm') }}</p>
        <div class="flex justify-end space-x-3 space-x-reverse">
          <button class="btn btn-secondary" @click="showDeleteProductConfirm=false">{{ $t('actions.cancel') }}</button>
          <button class="btn btn-danger" @click="deleteProductNow" :disabled="loading">{{ $t('actions.delete') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from "vue";
import { useI18n } from "vue-i18n";
import { PencilIcon } from "@heroicons/vue/24/outline";
import { useProductStore, type Product } from "../stores/products";

const { t } = useI18n();
const productStore = useProductStore();

// Reactive data
const products = ref<Product[]>([]);
const loading = ref(false);
const searchQuery = ref("");
const currentPage = ref(1);
const pageSize = ref(20);
const totalCount = ref(0);
const showCreateModal = ref(false);
const showEditModal = ref(false);
const currentProduct = ref({
  id: undefined as number | undefined,
  name: "",
  description: "",
  price: 0,
  sku: "",
});
const errorMessage = ref("");
const successMessage = ref("");

// Delete confirmation state
const showDeleteProductConfirm = ref(false);

function confirmDeleteProduct() {
  showDeleteProductConfirm.value = true;
}

async function deleteProductNow() {
  if (!currentProduct.value.id) return;
  try {
    loading.value = true;
    await productStore.deleteProduct(currentProduct.value.id);
    showDeleteProductConfirm.value = false;
    closeModal();
    loadProducts();
  } catch (e) {
    console.error(e);
    errorMessage.value = e instanceof Error ? e.message : t('products.delete_failed');
  } finally {
    loading.value = false;
  }
}

// Computed
const totalPages = computed(() => Math.ceil(totalCount.value / pageSize.value));

// Methods
async function loadProducts() {
  try {
    loading.value = true;
    const offset = (currentPage.value - 1) * pageSize.value;
    const result = await productStore.fetchProducts(
      searchQuery.value,
      pageSize.value,
      offset
    );
    products.value = result.data;
    totalCount.value = result.total;
  } catch (error) {
    console.error("Failed to load products:", error);
  } finally {
    loading.value = false;
  }
}

function debouncedSearch() {
  currentPage.value = 1;
  setTimeout(() => {
    loadProducts();
  }, 300);
}

function previousPage() {
  if (currentPage.value > 1) {
    currentPage.value--;
    loadProducts();
  }
}

function nextPage() {
  if (currentPage.value < totalPages.value) {
    currentPage.value++;
    loadProducts();
  }
}

function editProduct(product: Product) {
  currentProduct.value = {
    id: product.id,
    name: product.name,
    description: product.description || "",
    price: product.unit_price_cents / 100, // Convert cents to dollars
    sku: product.sku || "",
  };
  showEditModal.value = true;
}

function closeModal() {
  showCreateModal.value = false;
  showEditModal.value = false;
  errorMessage.value = "";
  successMessage.value = "";
  currentProduct.value = {
    id: undefined,
    name: "",
    description: "",
    price: 0,
    sku: "",
  };
}

async function saveProduct() {
  try {
    loading.value = true;
    errorMessage.value = "";
    successMessage.value = "";

    if (showCreateModal.value) {
      await productStore.createProduct(currentProduct.value);
      successMessage.value = t('messages.success.created');
    } else if (currentProduct.value.id) {
      await productStore.updateProduct({
        id: currentProduct.value.id,
        name: currentProduct.value.name,
        description: currentProduct.value.description,
        price: currentProduct.value.price,
        sku: currentProduct.value.sku,
      });
      successMessage.value = t('messages.success.updated');
    }

    closeModal();
    loadProducts();
  } catch (error) {
    console.error("Failed to save product:", error);
    errorMessage.value = error instanceof Error ? error.message : t('messages.error.validation_failed');
  } finally {
    loading.value = false;
  }
}

function formatPrice(priceCents: number): string {
  return new Intl.NumberFormat("ar-DZ", {
    style: "currency",
    currency: "DZD",
  }).format(priceCents / 100); // Convert cents to dinars
}

onMounted(() => {
  loadProducts();
});
</script>
