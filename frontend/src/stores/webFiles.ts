import { computed, ref } from "vue";
import { defineStore } from "pinia";
import { taskSuccess } from "@/composables/useAsyncTask";
import { useLatestAsyncTask } from "@/composables/useLatestAsyncTask";
import { fileApi, type FileEntry, type ListResult, type ServerStatus } from "@/lib/api";

export const useWebFilesStore = defineStore("webFiles", () => {
  const currentPath = ref("");
  const listing = ref<ListResult | null>(null);
  const status = ref<ServerStatus | null>(null);
  const { busy: loading, error, run, runLatest } = useLatestAsyncTask(true);
  const newFolderName = ref("");

  const canUpload = computed(() => status.value?.permission === "upload" || status.value?.permission === "manage");
  const canDelete = computed(() => status.value?.permission === "manage");
  const crumbs = computed(() => (currentPath.value ? currentPath.value.split("/").filter(Boolean) : []));
  const permissionLabel = computed(() => status.value?.permissions.find((item) => item.value === status.value?.permission)?.label ?? "只读");

  async function fetchListing(path: string) {
    const [nextStatus, nextListing] = await Promise.all([fileApi.status(), fileApi.list(path)]);
    status.value = nextStatus;
    listing.value = nextListing;
    currentPath.value = nextListing.path;
  }

  async function load(path = currentPath.value) {
    return runLatest(
      () => Promise.all([fileApi.status(), fileApi.list(path)]),
      ([nextStatus, nextListing]) => {
        status.value = nextStatus;
        listing.value = nextListing;
        currentPath.value = nextListing.path;
      },
    );
  }

  async function openEntry(entry: FileEntry) {
    if (entry.isDir) return load(entry.path);
    return taskSuccess(undefined);
  }

  async function uploadFiles(files: FileList | null) {
    if (!files?.length) return taskSuccess(false);
    return run(async () => {
      await fileApi.upload(currentPath.value, files);
      await fetchListing(currentPath.value);
      return true;
    });
  }

  async function deleteEntry(entry: FileEntry) {
    if (!window.confirm(`删除 ${entry.name}？文件会被移入 .lanfolder/trash。`)) return taskSuccess(false);
    return run(async () => {
      await fileApi.delete(entry.path);
      await fetchListing(currentPath.value);
      return true;
    });
  }

  async function createFolder() {
    if (!newFolderName.value.trim()) return taskSuccess(undefined);
    return run(async () => {
      await fileApi.mkdir(currentPath.value, newFolderName.value.trim());
      newFolderName.value = "";
      await fetchListing(currentPath.value);
    });
  }

  async function jumpToCrumb(index: number) {
    return load(crumbs.value.slice(0, index + 1).join("/"));
  }

  return {
    currentPath,
    listing,
    status,
    loading,
    error,
    newFolderName,
    canUpload,
    canDelete,
    crumbs,
    permissionLabel,
    load,
    openEntry,
    uploadFiles,
    deleteEntry,
    createFolder,
    jumpToCrumb,
  };
});
