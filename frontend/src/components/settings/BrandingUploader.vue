<template>
  <div class="sgd-branding-uploader">
    <h3>{{ $t("settings.brandingAssets") }}</h3>
    <p class="small">{{ $t("settings.brandingAssetsHelp") }}</p>
    <p v-if="loaded && !configured" class="small sgd-warning">
      {{ $t("settings.brandingNotConfigured") }}
    </p>

    <div class="sgd-branding-grid">
      <div v-for="item in items" :key="item.type" class="sgd-branding-item">
        <div class="sgd-preview">
          <img
            :src="previewURL(item.path)"
            :alt="$t(item.label)"
            @error="hideImage"
            @load="showImage"
          />
        </div>
        <div class="sgd-branding-info">
          <span class="sgd-branding-label">{{ $t(item.label) }}</span>
          <span class="sgd-status">
            {{
              status[item.type] ||
              (custom[item.type]
                ? $t("settings.brandingCustom")
                : $t("settings.brandingDefault"))
            }}
          </span>
          <div class="sgd-upload-actions">
            <input
              type="file"
              :id="'upload-' + item.type"
              :accept="item.accept"
              @change="handleFileUpload(item.type, $event)"
              style="display: none"
            />
            <button
              type="button"
              class="button button--flat"
              :disabled="!configured"
              @click="triggerFileInput(item.type)"
            >
              {{ $t("buttons.upload") }}
            </button>
            <button
              v-if="custom[item.type]"
              type="button"
              class="button button--flat button--red"
              @click="removeFile(item.type)"
            >
              {{ $t("buttons.delete") }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { fetchJSON, fetchURL } from "@/api/utils";
import { staticURL } from "@/utils/constants";
import { applyBrandingVersion, withBrandingVersion } from "@/utils/branding";

type BrandingStatus = {
  custom: Record<string, boolean>;
  version: string;
  configured: boolean;
};

const items = [
  {
    type: "logo",
    label: "settings.logo",
    path: "img/logo.svg",
    accept: ".svg,.png,.jpg,.jpeg",
  },
  {
    type: "favicon",
    label: "settings.favicon",
    path: "img/icons/favicon.svg",
    accept: ".svg,.ico,.png",
  },
  {
    type: "apple-touch-icon",
    label: "settings.appleTouchIcon",
    path: "img/icons/apple-touch-icon.png",
    accept: ".png",
  },
  {
    type: "android-chrome-192",
    label: "settings.androidIcon192",
    path: "img/icons/android-chrome-192x192.png",
    accept: ".png",
  },
  {
    type: "android-chrome-512",
    label: "settings.androidIcon512",
    path: "img/icons/android-chrome-512x512.png",
    accept: ".png",
  },
];

const { t } = useI18n();
const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const status = ref<Record<string, string>>({});
const custom = ref<Record<string, boolean>>({});
const version = ref(window.FileBrowser.BrandingVersion || "0");
const configured = ref(true);
const loaded = ref(false);

const previewURL = (path: string) =>
  withBrandingVersion(`${staticURL}/${path}`, version.value);

const hideImage = (event: Event) => {
  (event.target as HTMLImageElement).style.visibility = "hidden";
};

const showImage = (event: Event) => {
  (event.target as HTMLImageElement).style.visibility = "";
};

// Asks the server what is uploaded; a changed version also refreshes the
// logo and favicon already on the page.
const refresh = async () => {
  const data = await fetchJSON<BrandingStatus>("/api/branding");
  custom.value = data.custom;
  configured.value = data.configured;
  loaded.value = true;
  if (data.version !== version.value) {
    version.value = data.version;
    applyBrandingVersion(data.version);
  }
};

onMounted(() => {
  refresh().catch($showError);
});

const flash = (type: string, message: string) => {
  status.value[type] = message;
  setTimeout(() => {
    status.value[type] = "";
  }, 3000);
};

const triggerFileInput = (type: string) => {
  (
    document.getElementById("upload-" + type) as HTMLInputElement | null
  )?.click();
};

const handleFileUpload = async (type: string, event: Event) => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("file", file);
  formData.append("fileType", type);

  status.value[type] = t("settings.uploading");
  try {
    await fetchURL("/api/branding/upload", { method: "POST", body: formData });
    await refresh();
    flash(type, t("settings.uploadSuccess"));
    $showSuccess(t("settings.uploadSuccess"));
  } catch (e: any) {
    flash(type, t("settings.uploadFailed"));
    $showError(e);
  }

  target.value = "";
};

const removeFile = async (type: string) => {
  try {
    await fetchURL(`/api/branding/${type}`, { method: "DELETE" });
    await refresh();
  } catch (e: any) {
    $showError(e);
  }
};
</script>

<style scoped>
.sgd-branding-uploader {
  margin: 1em 0;
}

.sgd-branding-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 0.6em;
  margin-top: 0.75em;
}

.sgd-branding-item {
  display: flex;
  align-items: center;
  gap: 0.75em;
  padding: 0.6em;
  border: 1px solid var(--borderPrimary);
  border-radius: 8px;
}

.sgd-preview {
  flex: 0 0 44px;
  height: 44px;
  border-radius: 6px;
  background: var(--background);
  border: 1px solid var(--borderPrimary);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.sgd-preview img {
  max-width: 36px;
  max-height: 36px;
}

.sgd-branding-info {
  display: flex;
  flex-direction: column;
  gap: 0.2em;
  min-width: 0;
}

.sgd-branding-label {
  font-weight: 500;
  font-size: 13px;
}

.sgd-status {
  font-size: 11px;
  color: var(--textPrimary);
}

.sgd-upload-actions {
  display: flex;
  gap: 0.35em;
  margin-top: 0.2em;
}

.sgd-upload-actions .button {
  padding: 0.25em 0.6em;
  font-size: 11px;
}

.sgd-warning {
  color: #fbbf24 !important;
}
</style>
