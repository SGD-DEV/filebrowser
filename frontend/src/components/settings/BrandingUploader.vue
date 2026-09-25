<template>
  <div class="sgd-branding-uploader">
    <h3>{{ $t("settings.brandingAssets") }}</h3>
    <p class="small">{{ $t("settings.brandingAssetsHelp") }}</p>

    <div class="sgd-branding-grid">
      <div class="sgd-branding-item">
        <label>{{ $t("settings.logo") }}</label>
        <div class="sgd-upload-area">
          <input
            type="file"
            :id="'upload-logo'"
            accept=".svg,.png,.jpg,.jpeg"
            @change="handleFileUpload('logo', $event)"
            style="display: none"
          />
          <button
            type="button"
            class="button button--flat"
            @click="triggerFileInput('logo')"
          >
            {{ $t("buttons.upload") }}
          </button>
          <span v-if="uploadStatus.logo" class="sgd-status">{{
            uploadStatus.logo
          }}</span>
        </div>
      </div>

      <div class="sgd-branding-item">
        <label>{{ $t("settings.favicon") }}</label>
        <div class="sgd-upload-area">
          <input
            type="file"
            :id="'upload-favicon'"
            accept=".ico,.svg,.png"
            @change="handleFileUpload('favicon', $event)"
            style="display: none"
          />
          <button
            type="button"
            class="button button--flat"
            @click="triggerFileInput('favicon')"
          >
            {{ $t("buttons.upload") }}
          </button>
          <span v-if="uploadStatus.favicon" class="sgd-status">{{
            uploadStatus.favicon
          }}</span>
        </div>
      </div>

      <div class="sgd-branding-item">
        <label>{{ $t("settings.appleTouchIcon") }}</label>
        <div class="sgd-upload-area">
          <input
            type="file"
            :id="'upload-apple-touch-icon'"
            accept=".png"
            @change="handleFileUpload('apple-touch-icon', $event)"
            style="display: none"
          />
          <button
            type="button"
            class="button button--flat"
            @click="triggerFileInput('apple-touch-icon')"
          >
            {{ $t("buttons.upload") }}
          </button>
          <span v-if="uploadStatus['apple-touch-icon']" class="sgd-status">{{
            uploadStatus["apple-touch-icon"]
          }}</span>
        </div>
      </div>

      <div class="sgd-branding-item">
        <label>{{ $t("settings.androidIcon192") }}</label>
        <div class="sgd-upload-area">
          <input
            type="file"
            :id="'upload-android-chrome-192'"
            accept=".png"
            @change="handleFileUpload('android-chrome-192', $event)"
            style="display: none"
          />
          <button
            type="button"
            class="button button--flat"
            @click="triggerFileInput('android-chrome-192')"
          >
            {{ $t("buttons.upload") }}
          </button>
          <span v-if="uploadStatus['android-chrome-192']" class="sgd-status">{{
            uploadStatus["android-chrome-192"]
          }}</span>
        </div>
      </div>

      <div class="sgd-branding-item">
        <label>{{ $t("settings.androidIcon512") }}</label>
        <div class="sgd-upload-area">
          <input
            type="file"
            :id="'upload-android-chrome-512'"
            accept=".png"
            @change="handleFileUpload('android-chrome-512', $event)"
            style="display: none"
          />
          <button
            type="button"
            class="button button--flat"
            @click="triggerFileInput('android-chrome-512')"
          >
            {{ $t("buttons.upload") }}
          </button>
          <span v-if="uploadStatus['android-chrome-512']" class="sgd-status">{{
            uploadStatus["android-chrome-512"]
          }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { inject, ref } from "vue";
import { useI18n } from "vue-i18n";
import { fetchURL } from "@/api/utils";

const { t } = useI18n();
const $showError = inject<IToastError>("$showError")!;
const $showSuccess = inject<IToastSuccess>("$showSuccess")!;

const uploadStatus = ref<Record<string, string>>({});

const triggerFileInput = (fileType: string) => {
  const input = document.getElementById("upload-" + fileType) as HTMLInputElement;
  if (input) {
    input.click();
  }
};

const handleFileUpload = async (fileType: string, event: Event) => {
  const target = event.target as HTMLInputElement;
  const files = target.files;
  if (!files || files.length === 0) return;

  const file = files[0];
  const formData = new FormData();
  formData.append("file", file);
  formData.append("fileType", fileType);

  uploadStatus.value[fileType] = t("settings.uploading");

  try {
    await fetchURL("/api/branding/upload", {
      method: "POST",
      body: formData,
    });
    uploadStatus.value[fileType] = t("settings.uploadSuccess");
    $showSuccess(t("settings.uploadSuccess"));
    setTimeout(() => {
      uploadStatus.value[fileType] = "";
    }, 3000);
  } catch (e: any) {
    uploadStatus.value[fileType] = t("settings.uploadFailed");
    $showError(e);
    setTimeout(() => {
      uploadStatus.value[fileType] = "";
    }, 3000);
  }

  target.value = "";
};
</script>

<style scoped>
.sgd-branding-uploader {
  margin: 1em 0;
}

.sgd-branding-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1em;
  margin-top: 1em;
}

.sgd-branding-item {
  display: flex;
  flex-direction: column;
  gap: 0.5em;
}

.sgd-branding-item label {
  font-weight: 500;
  font-size: 0.9em;
}

.sgd-upload-area {
  display: flex;
  align-items: center;
  gap: 0.5em;
}

.sgd-status {
  font-size: 0.85em;
  color: var(--textPrimary);
}
</style>
